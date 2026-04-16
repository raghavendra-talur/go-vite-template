import { useState } from "react";
import { Copy, Check } from "lucide-react";

const examples = [
  {
    title: "Add a todo list",
    prompt:
      "Add a todo list feature to this app. Create a new todos module in server-go/modules/todos/ with CRUD endpoints and a SQLite migration. On the frontend, replace the landing page content with a clean todo list UI that lets me add, check off, and delete items. Use the existing Tailwind theme and design patterns.",
  },
  {
    title: "Add a dashboard",
    prompt:
      "Replace the landing page with a dashboard that shows system stats. Add a new stats API endpoint that returns the server uptime, database size, number of API tokens, and Go runtime memory stats. Display these in a grid of cards with the existing styling. Include a refresh button.",
  },
  {
    title: "Add dark mode toggle",
    prompt:
      "Add a dark mode toggle button to the top-right corner of the app. The CSS already supports .dark class on the html element — just add a button that toggles it and persists the preference in localStorage. Use a sun/moon icon from lucide-react.",
  },
];

function PromptCard({ title, prompt }: { title: string; prompt: string }) {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    await navigator.clipboard.writeText(prompt);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <button
      onClick={handleCopy}
      className="text-left w-full p-4 rounded-md transition-colors"
      style={{
        background: "var(--clr-surface)",
        border: "1px solid var(--clr-border-subtle)",
      }}
    >
      <div className="flex items-start justify-between gap-2 mb-2">
        <span
          className="text-[13px] font-semibold"
          style={{ color: "var(--clr-text-primary)" }}
        >
          {title}
        </span>
        {copied ? (
          <Check size={14} style={{ color: "var(--clr-green)", flexShrink: 0 }} />
        ) : (
          <Copy size={14} style={{ color: "var(--clr-text-muted)", flexShrink: 0 }} />
        )}
      </div>
      <p className="text-[11px] leading-relaxed" style={{ color: "var(--clr-text-muted)" }}>
        {prompt}
      </p>
    </button>
  );
}

export default function Landing() {
  return (
    <div className="flex-1 flex items-center justify-center p-8 overflow-auto">
      <div className="max-w-lg w-full space-y-8">
        <div className="text-center space-y-3">
          <h1
            className="text-lg font-semibold"
            style={{ color: "var(--clr-text-primary)" }}
          >
            __DISPLAY_NAME__
          </h1>
          <p className="text-[13px]" style={{ color: "var(--clr-text-secondary)" }}>
            Build this app with an AI agent — live.
          </p>
        </div>

        <div className="space-y-2" style={{ color: "var(--clr-text-muted)" }}>
          <p className="text-[12px]">
            The terminal below runs an AI coding agent in your repo.
            Ask it to build features — changes appear here instantly via hot reload.
            This page is yours to replace. Start by asking the agent to build something.
          </p>
        </div>

        <div className="space-y-3">
          <p
            className="text-[11px] uppercase tracking-wide"
            style={{ color: "var(--clr-text-muted)" }}
          >
            Example prompts — click to copy
          </p>
          {examples.map((ex) => (
            <PromptCard key={ex.title} {...ex} />
          ))}
        </div>

        <div className="text-center">
          <p className="text-[11px]" style={{ color: "var(--clr-text-muted)" }}>
            Press{" "}
            <span className="kbd-hint">Ctrl+`</span>{" "}
            to toggle the terminal
          </p>
        </div>
      </div>
    </div>
  );
}
