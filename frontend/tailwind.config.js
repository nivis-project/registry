/** @type {import('tailwindcss').Config} */
// Colours, families and radii come from src/styles/tokens.css. Components then
// write semantic utilities (bg-surface, text-ink, border-line) and almost never
// need a dark: variant, because the tokens already carry both themes.
export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  darkMode: ['selector', '[data-theme="dark"]'],
  theme: {
    extend: {
      colors: {
        ground: "var(--ground)",
        surface: "var(--surface)",
        ink: "var(--ink)",
        muted: "var(--muted)",
        line: "var(--line)",
        accent: "var(--accent)",
        "accent-soft": "var(--accent-soft)",
        warm: "var(--warm)",
        "on-warm": "var(--on-warm)",
        "band-bg": "var(--band-bg)",
        "band-ink": "var(--band-ink)",
        "band-muted": "var(--band-muted)",
        "band-line": "var(--band-line)",
        "code-bg": "var(--code-bg)",
        "code-ink": "var(--code-ink)",
        "code-dim": "var(--code-dim)",
        "tok-keyword": "var(--tok-keyword)",
        "tok-string": "var(--tok-string)",
        "tok-func": "var(--tok-func)",
        "tok-literal": "var(--tok-literal)",
        "mark-a": "var(--mark-a)",
        "mark-b": "var(--mark-b)",
        "mark-core": "var(--mark-core)",
        scrim: "var(--scrim)",
        "warning-bg": "var(--warning-bg)",
        "warning-ink": "var(--warning-ink)",
        "warning-line": "var(--warning-line)",
        "danger-bg": "var(--danger-bg)",
        "danger-ink": "var(--danger-ink)",
        "danger-line": "var(--danger-line)",
      },
      fontFamily: {
        sans: ["Hind", "Segoe UI", "system-ui", "sans-serif"],
        mono: ["IBM Plex Mono", "ui-monospace", "SFMono-Regular", "monospace"],
      },
      borderRadius: { card: "14px", box: "10px" },
      maxWidth: { shell: "1240px" },
      fontSize: { body: ["17px", "1.5"] },
    },
  },
  plugins: [],
};
