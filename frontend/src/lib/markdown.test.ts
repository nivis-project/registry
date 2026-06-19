import { describe, expect, it } from "vitest";
import { parseMarkdown } from "./markdown";

// A representative Nix-constructor document body (as tools/generate emits).
const nixDoc = `# \`random_password\`

Nivis Nix constructor for the \`random_password\` resource (provider \`random\`).

## Constructor

\`\`\`nix
{ name, length, keepers ? null, overrides ? {} }
\`\`\`

## Required arguments

- \`length\`

## Computed outputs

Read these from the resource's \`refAttr\`:

- \`result\`
`;

describe("parseMarkdown", () => {
  it("parses headings, code fences, lists, and paragraphs", () => {
    const blocks = parseMarkdown(nixDoc);
    const kinds = blocks.map((b) => b.kind);
    expect(kinds).toContain("heading");
    expect(kinds).toContain("code");
    expect(kinds).toContain("list");
    expect(kinds).toContain("para");
  });

  it("captures the nix language on the constructor code fence", () => {
    const code = parseMarkdown(nixDoc).find((b) => b.kind === "code");
    expect(code).toBeDefined();
    if (code && code.kind === "code") {
      expect(code.lang).toBe("nix");
      expect(code.text).toContain("overrides ? {}");
    }
  });

  it("never produces an HCL resource block from the Nix doc", () => {
    // The rendered doc body must be Nix, not HCL. No code block should contain
    // the canonical `resource "type" "name" {` HCL form.
    const codeBlocks = parseMarkdown(nixDoc).filter((b) => b.kind === "code");
    for (const b of codeBlocks) {
      if (b.kind === "code") {
        expect(b.text).not.toMatch(/^\s*resource\s+"/m);
      }
    }
  });
});
