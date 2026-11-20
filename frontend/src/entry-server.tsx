import { StrictMode } from "react";
import { renderToString } from "react-dom/server";
import { StaticRouter } from "react-router-dom/server";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { App } from "./App";
import { setTextFetcher, type TextFetcher } from "./lib/data";
import { headFor, headTags, type PageHead } from "./lib/head";
import { loadRoute } from "./lib/queries";

export { matchRoute } from "./lib/queries";

export interface RenderedPage {
  html: string;
  head: PageHead;
  headTags: string;
}

export async function render(pathname: string, fetcher: TextFetcher): Promise<RenderedPage> {
  setTextFetcher(fetcher);
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false, staleTime: Infinity } },
  });
  const route = await loadRoute(client, pathname);
  const html = renderToString(
    <StrictMode>
      <QueryClientProvider client={client}>
        <StaticRouter location={pathname}>
          <App />
        </StaticRouter>
      </QueryClientProvider>
    </StrictMode>,
  );
  const head = headFor(route, pathname, client);
  client.clear();
  return { html, head, headTags: headTags(head) };
}
