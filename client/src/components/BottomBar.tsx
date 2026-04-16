import { ChevronUp, ChevronDown } from "lucide-react";

export type AgentStatus = "connecting" | "running" | "stopped" | "error";

interface BottomBarProps {
  isOpen: boolean;
  onToggle: () => void;
  status: AgentStatus;
}

const statusConfig: Record<AgentStatus, { color: string; label: string }> = {
  connecting: { color: "var(--clr-orange)", label: "connecting" },
  running: { color: "var(--clr-green)", label: "agent running" },
  stopped: { color: "var(--clr-text-muted)", label: "agent stopped" },
  error: { color: "var(--clr-red)", label: "error" },
};

export default function BottomBar({ isOpen, onToggle, status }: BottomBarProps) {
  const { color, label } = statusConfig[status];
  const Icon = isOpen ? ChevronDown : ChevronUp;

  return (
    <div
      className="flex items-center justify-between px-4 select-none shrink-0"
      style={{
        height: 36,
        background: "var(--clr-surface-raised)",
        borderTop: "1px solid var(--clr-border-visible)",
        color: "var(--clr-text-secondary)",
        fontSize: 12,
      }}
    >
      <div className="flex items-center gap-2">
        <span
          style={{
            width: 7,
            height: 7,
            borderRadius: "50%",
            background: color,
            display: "inline-block",
          }}
        />
        <span>{label}</span>
      </div>
      <button
        onClick={onToggle}
        className="flex items-center gap-1 hover:opacity-80"
        style={{ color: "var(--clr-text-muted)", fontSize: 12 }}
      >
        <span>Terminal</span>
        <Icon size={14} />
        <span className="kbd-hint ml-1">Ctrl+`</span>
      </button>
    </div>
  );
}
