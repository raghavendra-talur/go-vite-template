import { useState } from "react";
import { setStoredToken } from "@/lib/queryClient";

interface TokenEntryProps {
  onAuthenticated: () => void;
}

export default function TokenEntry({ onAuthenticated }: TokenEntryProps) {
  const [token, setToken] = useState("");
  const [error, setError] = useState("");
  const [isValidating, setIsValidating] = useState(false);

  const handleSubmit = async () => {
    if (!token.trim()) return;
    setIsValidating(true);
    setError("");

    try {
      const res = await fetch("/api/v1/health", {
        headers: { Authorization: `Bearer ${token.trim()}` },
      });
      if (res.status === 401) {
        setError("Invalid token");
        return;
      }
      if (!res.ok) {
        setError("Connection error");
        return;
      }
      setStoredToken(token.trim());
      onAuthenticated();
    } catch {
      setError("Could not connect to server");
    } finally {
      setIsValidating(false);
    }
  };

  return (
    <div
      className="min-h-screen flex items-center justify-center p-4"
      style={{ background: "hsl(var(--background))", color: "hsl(var(--foreground))" }}
    >
      <div className="max-w-sm w-full space-y-6">
        <div className="text-center space-y-2">
          <h1 className="text-lg font-semibold">authenticate</h1>
          <p className="text-[12px]" style={{ color: "var(--clr-text-muted)" }}>
            Enter your API token. On the server machine, run{" "}
            <code className="text-[11px]" style={{ color: "var(--clr-text-secondary)" }}>
              ./dist/__APP_NAME__ admin-token --raw
            </code>{" "}
            or open <code className="text-[11px]" style={{ color: "var(--clr-text-secondary)" }}>DATA_DIR/adminAuth.json</code>.
          </p>
        </div>
        <div className="space-y-3">
          <input
            type="password"
            placeholder="rb_..."
            value={token}
            onChange={(e) => setToken(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && handleSubmit()}
            autoFocus
            className="w-full px-3 py-2 text-[13px] rounded border bg-transparent outline-none"
            style={{
              borderColor: "var(--clr-border-visible)",
              color: "var(--clr-text-primary)",
            }}
          />
          {error && (
            <p className="text-[12px]" style={{ color: "var(--clr-red)" }}>{error}</p>
          )}
          <button
            onClick={handleSubmit}
            disabled={isValidating || !token.trim()}
            className="w-full px-3 py-2 text-[13px] rounded border transition-colors"
            style={{
              borderColor: "var(--clr-border-visible)",
              color: "var(--clr-text-primary)",
              background: "var(--clr-surface-raised)",
              opacity: isValidating || !token.trim() ? 0.5 : 1,
            }}
          >
            {isValidating ? "validating..." : "connect"}
          </button>
        </div>
      </div>
    </div>
  );
}
