import { useState, useEffect, useCallback } from "react";
import { queryClient, getStoredToken, clearStoredToken, setOnUnauthorized } from "./lib/queryClient";
import { QueryClientProvider } from "@tanstack/react-query";
import TokenEntry from "@/pages/TokenEntry";
import Landing from "@/pages/Landing";
import Terminal from "@/components/Terminal";
import BottomBar, { type AgentStatus } from "@/components/BottomBar";

function MainApp() {
  const [terminalOpen, setTerminalOpen] = useState(true);
  const [agentStatus, setAgentStatus] = useState<AgentStatus>("connecting");

  useEffect(() => {
    const handleKey = (e: KeyboardEvent) => {
      if (e.ctrlKey && e.key === "`") {
        e.preventDefault();
        setTerminalOpen((prev) => !prev);
      }
    };
    window.addEventListener("keydown", handleKey);
    return () => window.removeEventListener("keydown", handleKey);
  }, []);

  return (
    <div
      className="flex flex-col overflow-hidden"
      style={{
        background: "hsl(var(--background))",
        height: "100dvh",
      }}
    >
      <Landing />
      <BottomBar
        isOpen={terminalOpen}
        onToggle={() => setTerminalOpen((prev) => !prev)}
        status={agentStatus}
      />
      {terminalOpen && (
        <div
          style={{
            height: "40vh",
            background: "#1a1a22",
            borderTop: "1px solid var(--clr-border-visible)",
            flexShrink: 0,
          }}
        >
          <Terminal onStatusChange={setAgentStatus} />
        </div>
      )}
    </div>
  );
}

function App() {
  const [needsAuth, setNeedsAuth] = useState(false);
  const [authChecked, setAuthChecked] = useState(false);

  const handleUnauthorized = useCallback(() => {
    clearStoredToken();
    setNeedsAuth(true);
  }, []);

  useEffect(() => {
    setOnUnauthorized(handleUnauthorized);

    const headers: Record<string, string> = {};
    const token = getStoredToken();
    if (token) {
      headers["Authorization"] = `Bearer ${token}`;
    }
    fetch("/api/v1/health", { headers })
      .then((res) => {
        if (res.status === 401) {
          setNeedsAuth(true);
        }
        setAuthChecked(true);
      })
      .catch(() => {
        setAuthChecked(true);
      });
  }, [handleUnauthorized]);

  if (!authChecked) return null;

  if (needsAuth) {
    return (
      <TokenEntry
        onAuthenticated={() => {
          setNeedsAuth(false);
          queryClient.invalidateQueries();
        }}
      />
    );
  }

  return (
    <QueryClientProvider client={queryClient}>
      <MainApp />
    </QueryClientProvider>
  );
}

export default App;
