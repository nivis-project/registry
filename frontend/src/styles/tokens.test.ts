import { describe, expect, it } from "vitest";
import { readdirSync, readFileSync, statSync } from "node:fs";
import { dirname, join, relative } from "node:path";
import { fileURLToPath } from "node:url";

// A rule nobody can keep by hand is not a rule. The briefing's acceptance
// criterion ("no default-palette utility and no colour literal outside the
// token file") is only true for as long as something checks it.

const SRC = dirname(dirname(fileURLToPath(import.meta.url)));
const TOKEN_FILE = "styles/tokens.css";

function sources(dir: string, out: string[] = []): string[] {
  for (const entry of readdirSync(dir)) {
    const full = join(dir, entry);
    if (statSync(full).isDirectory()) {
      if (entry === "fonts") continue; // vendored, not ours
      sources(full, out);
    } else if (/\.(tsx?|css)$/.test(entry) && !entry.endsWith(".test.ts") && !entry.endsWith(".test.tsx")) {
      out.push(full);
    }
  }
  return out;
}

// Tailwind's default families. `warm`, `ink`, `line` and friends are ours and
// must not match, so the palette names are listed explicitly.
const PALETTE =
  /\b(?:slate|gray|zinc|neutral|stone|red|orange|amber|yellow|lime|green|emerald|teal|cyan|blue|indigo|violet|purple|fuchsia|pink|rose)-(?:50|\d{3})\b/;
const LITERAL = /(?:#[0-9a-fA-F]{3,8}\b|\boklch\(|\brgb\(|\bhsl\()/;

describe("the token file is the only place a colour exists", () => {
  const files = sources(SRC);

  it("finds the application source", () => {
    expect(files.length).toBeGreaterThan(10);
  });

  it("has no default-palette utility outside the tokens", () => {
    const offenders: string[] = [];
    for (const f of files) {
      const rel = relative(SRC, f);
      if (rel === TOKEN_FILE) continue;
      readFileSync(f, "utf8")
        .split("\n")
        .forEach((line: string, i: number) => {
          const m = line.match(PALETTE);
          if (m) offenders.push(`${rel}:${i + 1} ${m[0]}`);
        });
    }
    expect(offenders, `use a semantic token instead:\n${offenders.join("\n")}`).toEqual([]);
  });

  it("has no colour literal outside the tokens", () => {
    const offenders: string[] = [];
    for (const f of files) {
      const rel = relative(SRC, f);
      if (rel === TOKEN_FILE) continue;
      readFileSync(f, "utf8")
        .split("\n")
        .forEach((line: string, i: number) => {
          const m = line.match(LITERAL);
          if (m) offenders.push(`${rel}:${i + 1} ${m[0]}`);
        });
    }
    expect(offenders, `define it in ${TOKEN_FILE}:\n${offenders.join("\n")}`).toEqual([]);
  });
});
