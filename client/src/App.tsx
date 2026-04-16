import { useState, useEffect, useCallback } from "react";
import { queryClient, getStoredToken, clearStoredToken, setOnUnauthorized } from "./lib/queryClient";
import { QueryClientProvider } from "@tanstack/react-query";
import TokenEntry from "@/pages/TokenEntry";

function MainApp() {
  return (
    <div
      className="flex flex-col overflow-hidden"
      style={{
        background: "hsl(var(--background))",
        height: "100dvh",
      }}
    >
      <div className="flex-1 flex items-center justify-center">
        <div className="text-center space-y-4">
          <h1 className="text-lg font-semibold" style={{ color: "var(--clr-text-primary)" }}>
            __DISPLAY_NAME__
          </h1>
          <p className="text-[13px]" style={{ color: "var(--clr-text-muted)" }}>
            Your app starts here. Edit <code className="text-[12px]" style={{ color: "var(--clr-text-secondary)" }}>client/src/App.tsx</code>.
          </p>
        </div>
      </div>
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
