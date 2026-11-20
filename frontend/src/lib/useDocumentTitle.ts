import { useEffect } from "react";
import { pageTitle } from "./head";

// A single-page app does not change the document title by itself, so a shared
// link or a browser-history entry would otherwise read the same for every page.
export function useDocumentTitle(title?: string) {
  useEffect(() => {
    document.title = pageTitle(title);
  }, [title]);
}
