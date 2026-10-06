import { describe, expect, it } from "vitest";
import { tokenizeNix } from "./nix";

const kinds = (src: string) =>
  tokenizeNix(src)
    .filter((t) => t.kind !== "text")
    .map((t) => [t.kind, t.text] as const);

describe("the Nix tokenizer", () => {
  it("recognises keywords", () => {
    expect(kinds("let x = 1; in x")).toContainEqual(["keyword", "let"]);
    expect(kinds("with pkgs; rec { }")).toContainEqual(["keyword", "with"]);
    expect(kinds("inherit (a) b;")).toContainEqual(["keyword", "inherit"]);
  });

  it("does not mistake a word containing a keyword for one", () => {
    // "letters" starts with "let"; "insecure" starts with "in".
    const k = kinds("{ letters = insecure; }");
    expect(k).not.toContainEqual(["keyword", "let"]);
    expect(k).not.toContainEqual(["keyword", "in"]);
  });

  it("takes strings whole, including a # inside one", () => {
    expect(kinds('x = "a # b";')).toContainEqual(["string", '"a # b"']);
  });

  it("handles indented strings", () => {
    expect(kinds("x = ''line # one'';")).toContainEqual(["string", "''line # one''"]);
  });

  it("takes comments to the end of the line, and block comments whole", () => {
    expect(kinds("a = 1; # trailing")).toContainEqual(["comment", "# trailing"]);
    expect(kinds("/* a\nb */ x")).toContainEqual(["comment", "/* a\nb */"]);
  });

  it("marks a dotted name applied to a set as a call", () => {
    expect(kinds("lib.mkProvider { }")).toContainEqual(["func", "lib.mkProvider"]);
  });

  it("does not mark a plain reference as a call", () => {
    expect(kinds("x = cfg.region;")).not.toContainEqual(["func", "cfg.region"]);
  });

  it("recognises literals", () => {
    expect(kinds("a = true; b = 42;")).toContainEqual(["literal", "true"]);
    expect(kinds("a = true; b = 42;")).toContainEqual(["literal", "42"]);
  });

  it("loses nothing: the tokens rejoin to the input", () => {
    const src = 'let a = "x"; # c\nin lib.f { b = 1; }';
    expect(tokenizeNix(src).map((t) => t.text).join("")).toBe(src);
  });
});
