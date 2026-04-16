import { useEffect, useRef, useCallback } from "react";
import { Terminal as XTerm } from "@xterm/xterm";
import { FitAddon } from "@xterm/addon-fit";
import { WebLinksAddon } from "@xterm/addon-web-links";
import "@xterm/xterm/css/xterm.css";
import { getStoredToken, getApiOrigin } from "@/lib/queryClient";

interface TerminalProps {
  onStatusChange?: (status: "connecting" | "running" | "stopped" | "error") => void;
}

export default function Terminal({ onStatusChange }: TerminalProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const termRef = useRef<XTerm | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const fitAddonRef = useRef<FitAddon | null>(null);

  const sendResize = useCallback(() => {
    const term = termRef.current;
    const ws = wsRef.current;
    if (term && ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type: "resize", cols: term.cols, rows: term.rows }));
    }
  }, []);

  useEffect(() => {
    if (!containerRef.current) return;

    const isDark = document.documentElement.classList.contains("dark");

    const term = new XTerm({
      fontFamily: "'JetBrains Mono', monospace",
      fontSize: 13,
      lineHeight: 1.4,
      cursorBlink: true,
      theme: isDark
        ? {
            background: "#0c0c14",
            foreground: "#e8e8f0",
            cursor: "#e8e8f0",
            selectionBackground: "rgba(255,255,255,0.12)",
          }
        : {
            background: "#1a1a22",
            foreground: "#e8e8f0",
            cursor: "#e8e8f0",
            selectionBackground: "rgba(255,255,255,0.12)",
          },
    });

    const fitAddon = new FitAddon();
    term.loadAddon(fitAddon);
    term.loadAddon(new WebLinksAddon());

    term.open(containerRef.current);
    fitAddon.fit();

    termRef.current = term;
    fitAddonRef.current = fitAddon;

    // WebSocket connection
    const origin = getApiOrigin();
    const wsProtocol = origin.startsWith("https") ? "wss" : "ws";
    const wsHost = origin.replace(/^https?:\/\//, "");
    const token = getStoredToken();
    const wsUrl = `${wsProtocol}://${wsHost}/api/v1/terminal/ws?token=${encodeURIComponent(token || "")}`;

    onStatusChange?.("connecting");
    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.binaryType = "arraybuffer";

    ws.onopen = () => {
      onStatusChange?.("running");
      sendResize();
    };

    ws.onmessage = (event) => {
      if (event.data instanceof ArrayBuffer) {
        term.write(new Uint8Array(event.data));
      } else {
        term.write(event.data);
      }
    };

    ws.onclose = () => {
      onStatusChange?.("stopped");
      term.write("\r\n\x1b[90m[session ended]\x1b[0m\r\n");
    };

    ws.onerror = () => {
      onStatusChange?.("error");
    };

    // Terminal input → WebSocket
    term.onData((data) => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(data);
      }
    });

    // Resize observer
    const observer = new ResizeObserver(() => {
      fitAddon.fit();
      sendResize();
    });
    observer.observe(containerRef.current);

    return () => {
      observer.disconnect();
      ws.close();
      term.dispose();
      termRef.current = null;
      wsRef.current = null;
      fitAddonRef.current = null;
    };
  }, [onStatusChange, sendResize]);

  return (
    <div
      ref={containerRef}
      className="w-full h-full"
      style={{ padding: "4px 0 0 8px" }}
    />
  );
}
