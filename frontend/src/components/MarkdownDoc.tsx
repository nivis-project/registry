import { Fragment, type ReactNode } from "react";
import { parseMarkdown } from "../lib/markdown";

// Render inline `code` spans within a text run (no other inline markup is
// emitted by our generator, so this stays deliberately small).
function inline(text: string): ReactNode[] {
  const out: ReactNode[] = [];
  const parts = text.split("`");
  parts.forEach((p, idx) => {
    if (idx % 2 === 1) {
      out.push(
        <code key={idx} className="rounded bg-slate-100 px-1 py-0.5 text-[0.9em] text-pink-700">
          {p}
        </code>,
      );
    } else if (p) {
      out.push(<Fragment key={idx}>{p}</Fragment>);
    }
  });
  return out;
}

// MarkdownDoc renders the Nix-constructor document body. The `nix` code fences
// carry the constructor signature and usage example; we label them so the page
// reads as a Nix reference, never HCL.
export function MarkdownDoc({ md }: { md: string }) {
  const blocks = parseMarkdown(md);
  return (
    <div className="space-y-4">
      {blocks.map((b, i) => {
        switch (b.kind) {
          case "heading": {
            const cls =
              b.level === 1
                ? "text-2xl font-bold text-slate-900"
                : b.level === 2
                  ? "mt-6 text-lg font-semibold text-slate-800"
                  : "mt-4 font-semibold text-slate-700";
            return (
              <div key={i} className={cls}>
                {inline(b.text)}
              </div>
            );
          }
          case "code":
            return (
              <pre
                key={i}
                data-lang={b.lang}
                className="overflow-x-auto rounded-lg bg-slate-900 p-4 text-sm text-slate-100"
              >
                {b.lang && (
                  <div className="mb-2 text-xs uppercase tracking-wide text-emerald-400">
                    {b.lang}
                  </div>
                )}
                <code>{b.text}</code>
              </pre>
            );
          case "list":
            return (
              <ul key={i} className="list-disc space-y-1 pl-6 text-slate-700">
                {b.items.map((it, j) => (
                  <li key={j}>{inline(it)}</li>
                ))}
              </ul>
            );
          case "para":
            return (
              <p key={i} className="text-slate-700">
                {inline(b.text)}
              </p>
            );
        }
      })}
    </div>
  );
}
