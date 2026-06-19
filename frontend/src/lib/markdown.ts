// A minimal, dependency-free Markdown block parser for the subset our generator
// emits: ATX headings (#..######), fenced code blocks (``` with optional lang),
// unordered list items (- ...), and paragraphs. This keeps the SPA light (no MDX
// pipeline) while rendering the Nix-constructor documents faithfully.
//
// The renderer (MarkdownDoc.tsx) turns these blocks into React elements; nothing
// here emits HTML, so there is no injection surface.

export type Block =
  | { kind: "heading"; level: number; text: string }
  | { kind: "code"; lang: string; text: string }
  | { kind: "list"; items: string[] }
  | { kind: "para"; text: string };

export function parseMarkdown(md: string): Block[] {
  const lines = md.replace(/\r\n/g, "\n").split("\n");
  const blocks: Block[] = [];
  let i = 0;

  const flushPara = (buf: string[]) => {
    const text = buf.join(" ").trim();
    if (text) blocks.push({ kind: "para", text });
  };

  let para: string[] = [];

  while (i < lines.length) {
    const line = lines[i];

    // Fenced code block.
    const fence = line.match(/^```(\w*)\s*$/);
    if (fence) {
      flushPara(para);
      para = [];
      const lang = fence[1] ?? "";
      const code: string[] = [];
      i++;
      while (i < lines.length && !/^```\s*$/.test(lines[i])) {
        code.push(lines[i]);
        i++;
      }
      i++; // skip closing fence
      blocks.push({ kind: "code", lang, text: code.join("\n") });
      continue;
    }

    // Heading.
    const heading = line.match(/^(#{1,6})\s+(.*)$/);
    if (heading) {
      flushPara(para);
      para = [];
      blocks.push({ kind: "heading", level: heading[1].length, text: heading[2].trim() });
      i++;
      continue;
    }

    // List (consecutive "- " lines).
    if (/^-\s+/.test(line)) {
      flushPara(para);
      para = [];
      const items: string[] = [];
      while (i < lines.length && /^-\s+/.test(lines[i])) {
        items.push(lines[i].replace(/^-\s+/, "").trim());
        i++;
      }
      blocks.push({ kind: "list", items });
      continue;
    }

    // Blank line ends a paragraph.
    if (line.trim() === "") {
      flushPara(para);
      para = [];
      i++;
      continue;
    }

    para.push(line.trim());
    i++;
  }
  flushPara(para);
  return blocks;
}
