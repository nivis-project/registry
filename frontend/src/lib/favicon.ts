import { markPaths } from "../components/Mark";

// The favicon is drawn from the same geometry as the header mark, so the tab
// icon cannot drift from the site. Colours are read from the tokens at call
// time, which also makes the icon follow the theme.
export function installFavicon() {
  const styles = getComputedStyle(document.documentElement);
  const fills = ["--mark-a", "--mark-b", "--mark-core"].map((v) =>
    styles.getPropertyValue(v).trim(),
  );

  const paths = markPaths
    .map((p, i) => `<path d="${p.d}" fill="${fills[i]}" opacity="${p.opacity}"/>`)
    .join("");
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="-100 -100 200 200">${paths}</svg>`;

  let link = document.querySelector<HTMLLinkElement>('link[rel="icon"]');
  if (!link) {
    link = document.createElement("link");
    link.rel = "icon";
    document.head.appendChild(link);
  }
  link.type = "image/svg+xml";
  link.href = `data:image/svg+xml,${encodeURIComponent(svg)}`;
}
