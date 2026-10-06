import { Fragment, type ReactNode } from "react";
import { parseMarkdown } from "../lib/markdown";
import { CodeBlock } from "./CodeBlock";

// Render inline `code` spans within a text run (no other inline markup is
// emitted by our generator, so this stays deliberately small).
function inline(text: string): ReactNode[] {
  const out: ReactNode[] = [];
  const parts = text.split("`");
  parts.forEach((p, idx) => {
    if (idx % 2 === 1) {
      out.push(
        <code
          key={idx}
          className="rounded bg-accent-soft px-1.5 py-0.5 font-mono text-[0.9em] text-ink"
        >
          {p}
        </code>,
      );
    } else if (p) {
      out.push(<Fragment key={idx}>{p}</Fragment>);
    }
  });
  return out;
}

// MarkdownDoc renders the Nix-constructor document body. The parser emits four
// block kinds and no more (heading, code, list, para), so this styles exactly
// those; it does not try to read meaning out of the prose.
export function MarkdownDoc({ md }: { md: string }) {
  const blocks = parseMarkdown(md);
  const headingSize = (level: number) =>
    level <= 1
      ? "text-[clamp(28px,3.4vw,40px)] font-medium font-mono text-ink"
      : level === 2
        ? "mt-8 text-[22px] font-semibold text-ink"
        : "mt-6 text-[17px] font-semibold text-ink";

  return (
    <div className="space-y-4">
      {blocks.map((b, i) => {
        switch (b.kind) {
          case "heading":
            return (
              <div key={i} className={headingSize(b.level)}>
                {inline(b.text)}
              </div>
            );
          case "code":
            return <CodeBlock key={i} code={b.text} lang={b.lang} />;
          case "list":
            return (
              <ul key={i} className="list-disc space-y-1 pl-6 text-ink">
                {b.items.map((it, j) => (
                  <li key={j}>{inline(it)}</li>
                ))}
              </ul>
            );
          case "para":
            return (
              <p key={i} className="text-ink">
                {inline(b.text)}
              </p>
            );
          default:
            return null;
        }
      })}
    </div>
  );
}
