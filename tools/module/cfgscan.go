package module

import (
	"regexp"
	"sort"
	"strings"
)

// CfgKey is one configuration path a module reads.
type CfgKey struct {
	Path     string `json:"path"`
	Required bool   `json:"required"`
}

// StripComments removes Nix comments so prose cannot be mistaken for code.
// nivis-aws-amplify-site's header comment names cfg.githubTokenParam in
// English; without this, a module that discusses a key it does not read would
// have that key invented for it.
//
// String literals are tracked so a "#" or "/*" inside one is left alone.
func StripComments(src string) string {
	var out strings.Builder
	out.Grow(len(src))
	const (
		code = iota
		lineComment
		blockComment
		dquote
		indented // Nix '' ... '' string
	)
	state := code
	for i := 0; i < len(src); i++ {
		c := src[i]
		switch state {
		case code:
			switch {
			case c == '#':
				state = lineComment
			case c == '/' && i+1 < len(src) && src[i+1] == '*':
				state = blockComment
				i++
			case c == '"':
				state = dquote
				out.WriteByte(c)
			case c == '\'' && i+1 < len(src) && src[i+1] == '\'':
				state = indented
				out.WriteString("''")
				i++
			default:
				out.WriteByte(c)
			}
		case lineComment:
			if c == '\n' {
				state = code
				out.WriteByte(c)
			}
		case blockComment:
			if c == '*' && i+1 < len(src) && src[i+1] == '/' {
				state = code
				i++
			}
		case dquote:
			out.WriteByte(c)
			if c == '\\' && i+1 < len(src) {
				i++
				out.WriteByte(src[i])
			} else if c == '"' {
				state = code
			}
		case indented:
			out.WriteByte(c)
			if c == '\'' && i+1 < len(src) && src[i+1] == '\'' {
				out.WriteByte(src[i+1])
				i++
				state = code
			}
		}
	}
	return out.String()
}

var (
	cfgRe      = regexp.MustCompile(`\bcfg((?:\.[A-Za-z_][A-Za-z0-9_'-]*)+)`)
	cfgOrRe    = regexp.MustCompile(`\bcfg((?:\.[A-Za-z_][A-Za-z0-9_'-]*)+)\s+or\b`)
	trailingOr = "or"
)

// ScanCfg reports the configuration paths a module reads and which of them are
// optional.
//
// This is a source scan, not a derivation: it is the only way to see an
// optional key, because `cfg.x or default` never fails and so is invisible to
// evaluation. Callers MUST label the result as inferred.
//
// Under-reporting (a dynamically computed attribute name) is acceptable.
// Over-reporting is not, which is why comments are stripped first.
func ScanCfg(src string) []CfgKey {
	src = StripComments(src)

	optional := map[string]bool{}
	for _, m := range cfgOrRe.FindAllStringSubmatch(src, -1) {
		optional[strings.TrimPrefix(m[1], ".")] = true
	}

	paths := map[string]bool{}
	for _, m := range cfgRe.FindAllStringSubmatch(src, -1) {
		paths[strings.TrimPrefix(m[1], ".")] = true
	}

	out := make([]CfgKey, 0, len(paths))
	for p := range paths {
		out = append(out, CfgKey{Path: p, Required: !optionalPath(p, optional)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}

// optionalPath reports whether a path is optional, either directly or because
// an ancestor is: a key under `cfg.formProxy or null` is only ever read when
// formProxy was supplied, so calling it required would be wrong.
func optionalPath(path string, optional map[string]bool) bool {
	parts := strings.Split(path, ".")
	for i := range parts {
		if optional[strings.Join(parts[:i+1], ".")] {
			return true
		}
	}
	return false
}
