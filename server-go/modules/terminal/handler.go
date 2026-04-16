package terminal

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/creack/pty"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"

	"__MODULE_PATH__/server-go/config"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type resizeMsg struct {
	Type string `json:"type"`
	Cols uint16 `json:"cols"`
	Rows uint16 `json:"rows"`
}

func RegisterRoutes(r chi.Router, cfg *config.Config) {
	r.Get("/terminal/ws", handleWS(cfg))
}

func handleWS(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("terminal: websocket upgrade failed: %v", err)
			return
		}
		defer conn.Close()

		if cfg.AgentCmd == "" {
			conn.WriteMessage(websocket.TextMessage, []byte(
				"No agent configured.\r\n\r\n"+
					"Set AGENT_CMD in your .env file (e.g., AGENT_CMD=claude) and restart the server.\r\n"))
			return
		}

		args := strings.Fields(cfg.AgentCmd)
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = cfg.WorkDir
		cmd.Env = append(os.Environ(), "TERM=xterm-256color")

		ptmx, err := pty.Start(cmd)
		if err != nil {
			log.Printf("terminal: pty start failed: %v", err)
			conn.WriteMessage(websocket.TextMessage, []byte(
				"Failed to start agent: "+err.Error()+"\r\n"))
			return
		}

		var once sync.Once
		cleanup := func() {
			once.Do(func() {
				ptmx.Close()
				if cmd.Process != nil {
					cmd.Process.Kill()
					cmd.Wait()
				}
			})
		}
		defer cleanup()

		// PTY → WebSocket
		go func() {
			buf := make([]byte, 4096)
			for {
				n, err := ptmx.Read(buf)
				if n > 0 {
					if wErr := conn.WriteMessage(websocket.BinaryMessage, buf[:n]); wErr != nil {
						break
					}
				}
				if err != nil {
					break
				}
			}
			cleanup()
			conn.Close()
		}()

		// WebSocket → PTY
		for {
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				break
			}
			if msgType == websocket.TextMessage {
				// Check if it's a resize message
				var resize resizeMsg
				if json.Unmarshal(msg, &resize) == nil && resize.Type == "resize" {
					pty.Setsize(ptmx, &pty.Winsize{
						Rows: resize.Rows,
						Cols: resize.Cols,
					})
					continue
				}
			}
			// Regular input — write to PTY
			if _, err := io.Copy(ptmx, strings.NewReader(string(msg))); err != nil {
				break
			}
		}

		cleanup()
		log.Printf("terminal: session closed")
	}
}
