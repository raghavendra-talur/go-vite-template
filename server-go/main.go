package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"strings"
	"syscall"

	"github.com/go-chi/chi/v5"
	"__MODULE_PATH__/server-go/config"
	"__MODULE_PATH__/server-go/db"
	"__MODULE_PATH__/server-go/middleware"
	"__MODULE_PATH__/server-go/modules/health"
	"__MODULE_PATH__/server-go/modules/terminal"
	"__MODULE_PATH__/server-go/modules/tokens"
)

//go:embed dist/public
var frontendFS embed.FS

// Version is set at build time via -ldflags.
var Version = "dev"

// lockFile holds the open lock file descriptor for the lifetime of the process.
var lockFile *os.File

func acquireDataDirLock(dataDir string) error {
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return err
	}
	f, err := os.OpenFile(dataDir+"/"+config.AppName+".lock", os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		return err
	}
	lockFile = f
	return nil
}

func tryBindAll(hosts []string, port int) ([]net.Listener, error) {
	var listeners []net.Listener
	for _, host := range hosts {
		ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
		if err != nil {
			for _, prev := range listeners {
				prev.Close()
			}
			return nil, err
		}
		listeners = append(listeners, ln)
	}
	return listeners, nil
}

func bindPorts(hosts []string, port int, explicit bool) ([]net.Listener, error) {
	if explicit {
		lns, err := tryBindAll(hosts, port)
		if err != nil {
			return nil, fmt.Errorf("cannot bind to port %d: %w", port, err)
		}
		return lns, nil
	}
	for p := port; p < port+100; p++ {
		lns, err := tryBindAll(hosts, p)
		if err == nil {
			if p != port {
				log.Printf("Port %d unavailable, using %d", port, p)
			}
			return lns, nil
		}
	}
	return nil, fmt.Errorf("no free port found in range %d-%d", port, port+99)
}

func printVersion() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		fmt.Printf("%s (unknown version)\n", config.AppName)
		return
	}

	vcs, rev, dirty, buildTime := "", "", false, ""
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs":
			vcs = s.Value
		case "vcs.revision":
			rev = s.Value
		case "vcs.modified":
			dirty = s.Value == "true"
		case "vcs.time":
			buildTime = s.Value
		}
	}

	version := Version
	if version == "dev" && rev != "" {
		short := rev
		if len(short) > 12 {
			short = short[:12]
		}
		version = short
	}
	if dirty {
		version += "-dirty"
	}

	fmt.Printf("%s %s\n", config.AppName, version)
	if vcs != "" {
		fmt.Printf("  vcs:      %s\n", vcs)
	}
	if rev != "" {
		fmt.Printf("  commit:   %s\n", rev)
	}
	if buildTime != "" {
		fmt.Printf("  built:    %s\n", buildTime)
	}
	fmt.Printf("  go:       %s\n", info.GoVersion)
}

func printAdminToken() {
	fs := flag.NewFlagSet("admin-token", flag.ExitOnError)
	dataDirFlag := fs.String("data-dir", "", "data directory (overrides DATA_DIR env var)")
	raw := fs.Bool("raw", false, "print token only")
	fs.Parse(os.Args[2:])

	cfg := config.Load()
	if *dataDirFlag != "" {
		cfg.DataDir = *dataDirFlag
	}

	logWriter := log.Writer()
	log.SetOutput(io.Discard)
	defer log.SetOutput(logWriter)

	token := config.EnsureAdminAuth(cfg.DataDir)
	path := filepath.Join(cfg.DataDir, "adminAuth.json")

	if *raw {
		fmt.Println(token)
		return
	}

	fmt.Printf("Admin token: %s\n", token)
	fmt.Printf("Token file:  %s\n", path)
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version":
			printVersion()
			return
		case "admin-token":
			printAdminToken()
			return
		}
	}

	portFlag := flag.Int("port", 0, "server listen port (overrides PORT env var, default 9002)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `%s — Go + Vite full-stack app

Usage:
  %s [flags]
  %s version
  %s admin-token [--raw] [--data-dir PATH]

Environment variables:
  PORT          Server listen port (default: 9002)
  HOST          Comma-separated bind addresses (default: 127.0.0.1)
  DATA_DIR      Base directory for database and files
  DB_PATH       SQLite database path (default: DATA_DIR/%s.db)

Flags:
`, config.AppName, config.AppName, config.AppName, config.AppName, config.AppName)
		flag.PrintDefaults()
	}
	flag.Parse()

	cfg := config.Load()

	if err := acquireDataDirLock(cfg.DataDir); err != nil {
		log.Fatalf("Another instance is already running with DATA_DIR=%s: %v", cfg.DataDir, err)
	}

	portExplicit := *portFlag != 0 || os.Getenv("PORT") != ""
	if *portFlag != 0 {
		cfg.Port = *portFlag
	}

	// Load or create admin auth token
	adminToken := config.EnsureAdminAuth(cfg.DataDir)
	middleware.SetAdminToken(adminToken)

	// Connect to SQLite database
	if err := db.Connect(cfg.DBPath); err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Run database migrations
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := db.Migrate(); err != nil {
		log.Fatal(err)
	}

	// Expose version to the health API
	health.AppVersion = Version

	// Setup router
	r := chi.NewRouter()
	r.Use(middleware.Logging)
	r.Use(middleware.CORS)

	// Health check — unauthenticated, for service monitors
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Register API modules under /api/v1
	r.Route("/api/v1", func(v1 chi.Router) {
		v1.Use(middleware.BearerAuth)

		health.RegisterRoutes(v1)
		tokens.RegisterRoutes(v1)
		terminal.RegisterRoutes(v1, cfg)

		// Register your modules here:
		// items.RegisterRoutes(v1)
	})

	// Serve frontend
	if cfg.DevMode {
		viteURL, err := url.Parse(cfg.ViteURL)
		if err != nil {
			log.Fatalf("Invalid VITE_URL: %v", err)
		}
		proxy := httputil.NewSingleHostReverseProxy(viteURL)
		r.NotFound(proxy.ServeHTTP)
		log.Printf("Dev mode: proxying frontend to %s", cfg.ViteURL)
	} else {
		staticFS, err := fs.Sub(frontendFS, "dist/public")
		if err != nil {
			log.Fatal(err)
		}
		fileServer := http.FileServer(http.FS(staticFS))
		r.NotFound(func(w http.ResponseWriter, r *http.Request) {
			path := strings.TrimPrefix(r.URL.Path, "/")
			if _, err := fs.Stat(staticFS, path); err != nil {
				// SPA fallback: serve index.html
				indexFile, _ := fs.ReadFile(staticFS, "index.html")
				w.Header().Set("Content-Type", "text/html")
				w.Write(indexFile)
				return
			}
			fileServer.ServeHTTP(w, r)
		})
	}

	listeners, err := bindPorts(cfg.Hosts, cfg.Port, portExplicit)
	if err != nil {
		log.Fatal(err)
	}

	for _, ln := range listeners {
		log.Printf("Server listening on %s", ln.Addr())
	}

	server := &http.Server{Handler: r}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down...")
		cancel()
		server.Close()
	}()

	for _, ln := range listeners[1:] {
		go func(l net.Listener) {
			if err := server.Serve(l); err != http.ErrServerClosed {
				log.Fatal(err)
			}
		}(ln)
	}
	if err := server.Serve(listeners[0]); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
