import { TOKEN_CLASS, tokenizeNix } from "../lib/nix";

// Code blocks are dark in both themes, by design: generated Nix reads as code,
// not as part of the page.
export function CodeBlock({ code, lang }: { code: string; lang?: string }) {
  const highlight = !lang || lang === "nix";
  return (
    <pre className="overflow-x-auto rounded-box bg-code-bg p-4 font-mono text-[15px] leading-relaxed text-code-ink">
      <code>
        {highlight
          ? tokenizeNix(code).map((t, i) => (
              <span key={i} className={TOKEN_CLASS[t.kind]}>
                {t.text}
              </span>
            ))
          : code}
      </code>
    </pre>
  );
}
