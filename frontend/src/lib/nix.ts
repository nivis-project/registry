// A small Nix tokenizer for the generated code blocks.
//
// The Markdown parser is deliberately dependency-free and the code it renders
// is generated Nix, a narrow and predictable subset. A highlighting library
// would cost more bundle than the whole stylesheet.

export type TokenKind = "keyword" | "string" | "comment" | "literal" | "func" | "text";
export interface Token {
  kind: TokenKind;
  text: string;
}

const KEYWORDS = new Set([
  "let", "in", "inherit", "with", "rec", "if", "then", "else", "assert",
]);
const LITERALS = new Set(["true", "false", "null"]);

const IDENT = /[A-Za-z_][A-Za-z0-9_'-]*(?:\.[A-Za-z_][A-Za-z0-9_'-]*)*/y;

export function tokenizeNix(src: string): Token[] {
  const out: Token[] = [];
  let buf = "";
  const flush = () => {
    if (buf) {
      out.push({ kind: "text", text: buf });
      buf = "";
    }
  };
  const push = (kind: TokenKind, text: string) => {
    flush();
    out.push({ kind, text });
  };

  let i = 0;
  while (i < src.length) {
    const c = src[i];

    // Comments: # to end of line, and /* ... */
    if (c === "#") {
      const end = src.indexOf("\n", i);
      const stop = end === -1 ? src.length : end;
      push("comment", src.slice(i, stop));
      i = stop;
      continue;
    }
    if (c === "/" && src[i + 1] === "*") {
      const end = src.indexOf("*/", i + 2);
      const stop = end === -1 ? src.length : end + 2;
      push("comment", src.slice(i, stop));
      i = stop;
      continue;
    }

    // Strings: "..." with escapes, and '' ... '' blocks.
    if (c === '"') {
      let j = i + 1;
      while (j < src.length && src[j] !== '"') j += src[j] === "\\" ? 2 : 1;
      push("string", src.slice(i, Math.min(j + 1, src.length)));
      i = j + 1;
      continue;
    }
    if (c === "'" && src[i + 1] === "'") {
      const end = src.indexOf("''", i + 2);
      const stop = end === -1 ? src.length : end + 2;
      push("string", src.slice(i, stop));
      i = stop;
      continue;
    }

    // Numbers.
    if (c >= "0" && c <= "9") {
      let j = i;
      while (j < src.length && /[0-9.]/.test(src[j])) j++;
      push("literal", src.slice(i, j));
      i = j;
      continue;
    }

    // Identifiers: keywords, literals, and dotted names applied to a set or a
    // string (lib.mkProvider { ... }), which is what a call looks like here.
    IDENT.lastIndex = i;
    const m = IDENT.exec(src);
    if (m && m.index === i) {
      const word = m[0];
      let j = i + word.length;
      if (KEYWORDS.has(word)) push("keyword", word);
      else if (LITERALS.has(word)) push("literal", word);
      else {
        let k = j;
        while (k < src.length && (src[k] === " " || src[k] === "\t")) k++;
        if (src[k] === "{" || src[k] === '"') push("func", word);
        else buf += word;
      }
      i = j;
      continue;
    }

    buf += c;
    i++;
  }
  flush();
  return out;
}

export const TOKEN_CLASS: Record<TokenKind, string> = {
  keyword: "text-tok-keyword",
  string: "text-tok-string",
  comment: "text-code-dim",
  literal: "text-tok-literal",
  func: "text-tok-func",
  text: "",
};
