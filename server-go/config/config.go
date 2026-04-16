package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

func init() {
	// Load .env from project root (one level up from server-go/)
	godotenv.Load(filepath.Join("..", ".env"))
	// Also try current directory for when running the built binary
	godotenv.Load(".env")
}

type Config struct {
	DBPath  string   // SQLite database file path
	Port    int
	Hosts   []string // bind addresses (default: ["127.0.0.1"])
	DataDir string   // local filesystem storage
	DevMode bool
	ViteURL string   // Vite dev server URL for proxying in dev mode
}

// expandHome replaces a leading ~ with the user's home directory.
func expandHome(path string) string {
	if path == "~" || path == "~/" {
		home, _ := os.UserHomeDir()
		return home
	}
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

// AppName is used for default data directory paths. Change this to your app name.
const AppName = "__APP_NAME__"

func Load() *Config {
	port := 9002
	if p := os.Getenv("PORT"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			port = v
		}
	}

	dataDir := expandHome(os.Getenv("DATA_DIR"))
	if dataDir == "" {
		xdgData := os.Getenv("XDG_DATA_HOME")
		if xdgData == "" {
			home, _ := os.UserHomeDir()
			xdgData = filepath.Join(home, ".local", "share")
		}
		dataDir = filepath.Join(xdgData, AppName)
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = filepath.Join(dataDir, AppName+".db")
	}

	hosts := []string{"127.0.0.1"}
	if h := os.Getenv("HOST"); h != "" {
		hosts = nil
		for _, addr := range strings.Split(h, ",") {
			addr = strings.TrimSpace(addr)
			if addr != "" {
				hosts = append(hosts, addr)
			}
		}
	}

	devMode := os.Getenv("NODE_ENV") == "development" || os.Getenv("DEV") == "1"

	viteURL := os.Getenv("VITE_URL")
	if viteURL == "" {
		viteURL = "http://127.0.0.1:9005"
	}

	return &Config{
		DBPath:  dbPath,
		Port:    port,
		Hosts:   hosts,
		DataDir: dataDir,
		DevMode: devMode,
		ViteURL: viteURL,
	}
}
