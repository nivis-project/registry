package generate

import (
	"fmt"
	"strings"
)

// RenderDoc renders a parsed Constructor as the per-item document body: the
// Nivis Nix reference for the resource. It shows the constructor signature, the
// required and optional arguments, the computed outputs (read via refAttr), and
// a usage example — all in Nix form. It deliberately emits NO HCL
// `resource "..." {}` block: the schema-derived Nix constructor IS the reference.
func RenderDoc(c Constructor) string {
	var b strings.Builder

	title := c.Type
	fmt.Fprintf(&b, "# `%s`\n\n", title)
	fmt.Fprintf(&b, "Nivis Nix constructor for the `%s` resource", c.Type)
	if c.Provider != "" {
		fmt.Fprintf(&b, " (provider `%s`)", c.Provider)
	}
	b.WriteString(". This reference is derived from the provider binary's own schema via `nivis gen` — it always matches the real provider.\n\n")

	// Signature (Nix lambda).
	b.WriteString("## Constructor\n\n")
	b.WriteString("```nix\n")
	b.WriteString(signature(c))
	b.WriteString("\n```\n\n")

	// Required arguments.
	b.WriteString("## Required arguments\n\n")
	if len(c.Required) == 0 {
		b.WriteString("_None._\n\n")
	} else {
		for _, a := range c.Required {
			fmt.Fprintf(&b, "- `%s`\n", a)
		}
		b.WriteString("\n")
	}

	// Optional arguments.
	b.WriteString("## Optional arguments\n\n")
	if len(c.Optional) == 0 {
		b.WriteString("_None._\n\n")
	} else {
		for _, a := range c.Optional {
			fmt.Fprintf(&b, "- `%s` (defaults to `null`)\n", a)
		}
		b.WriteString("\n")
	}

	// Computed outputs.
	b.WriteString("## Computed outputs\n\n")
	if len(c.Computed) == 0 {
		b.WriteString("_None._\n\n")
	} else {
		b.WriteString("Read these from the resource's `refAttr`:\n\n")
		for _, o := range c.Computed {
			fmt.Fprintf(&b, "- `%s`\n", o)
		}
		b.WriteString("\n")
	}

	// Usage example (Nix).
	b.WriteString("## Example\n\n")
	b.WriteString("```nix\n")
	b.WriteString(example(c))
	b.WriteString("\n```\n")

	return b.String()
}

// signature renders the constructor lambda argument set in Nix.
// sigWrapWidth is the column past which a one-line constructor signature is
// broken to one argument per line. Roughly a comfortable code-block width so the
// signature stays readable when a resource has many (often mandatory) arguments.
const sigWrapWidth = 72

// signature renders the constructor's argument set. It is width-aware: a short
// signature stays on one line; a wide one breaks to one argument per line —
//
//	{
//	  name,
//	  arg1,
//	  arg2 ? null,
//	  overrides ? {}
//	}
//
// so a resource with many mandatory arguments stays readable.
func signature(c Constructor) string {
	parts := []string{"name"}
	for _, a := range c.Required {
		parts = append(parts, a)
	}
	for _, a := range c.Optional {
		parts = append(parts, a+" ? null")
	}
	parts = append(parts, "overrides ? {}")

	oneLine := "{ " + strings.Join(parts, ", ") + " }"
	if len(oneLine) <= sigWrapWidth {
		return oneLine
	}
	// Wrap: one argument per line, two-space indented, trailing-comma-free last.
	var b strings.Builder
	b.WriteString("{\n")
	for i, p := range parts {
		b.WriteString("  ")
		b.WriteString(p)
		if i < len(parts)-1 {
			b.WriteString(",")
		}
		b.WriteString("\n")
	}
	b.WriteString("}")
	return b.String()
}

// example renders a minimal instantiation: name + each required argument.
func example(c Constructor) string {
	var b strings.Builder
	resName := exampleName(c.Type)
	fmt.Fprintf(&b, "%s {\n", c.Type)
	fmt.Fprintf(&b, "  name = \"%s\";\n", resName)
	for _, a := range c.Required {
		fmt.Fprintf(&b, "  %s = …;  # required\n", a)
	}
	b.WriteString("}")
	return b.String()
}

// exampleName derives a short instance name from a type ("random_password" ->
// "password", "tls_private_key" -> "private_key").
func exampleName(typ string) string {
	if i := strings.IndexByte(typ, '_'); i >= 0 && i+1 < len(typ) {
		return typ[i+1:]
	}
	return typ
}

// ContainsHCLResourceBlock reports whether s contains an HCL `resource "..." {}`
// reference block — used by tests to assert the rendered doc is Nix, not HCL.
func ContainsHCLResourceBlock(s string) bool {
	// Look for the canonical HCL form: resource "type" "name" {
	for _, line := range strings.Split(s, "\n") {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "resource \"") && strings.HasSuffix(t, "{") {
			return true
		}
	}
	return false
}
