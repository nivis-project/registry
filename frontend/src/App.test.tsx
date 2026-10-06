import { afterEach, describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { App } from "./App";
import { sections } from "./components/SiteNav";

const catalog = [
  {
    namespace: "hashicorp",
    name: "random",
    version: "3.9.0",
    tier: "compatible by design",
  },
];

const providerIndex = {
  id: "3.9.0",
  docs: {
    resources: [{ name: "random_password", title: "random_password" }],
    datasources: [],
    functions: [],
    guides: [],
  },
  compat: {
    address: "hashicorp/random",
    tier: "compatible by design",
    schema_extractable: true,
    protocols: ["5.0"],
    architectures: ["linux/amd64"],
    e2e: "none",
  },
};

function stubContract() {
  vi.stubGlobal(
    "fetch",
    vi.fn(async (url: string) => {
      if (url.endsWith("catalog.json"))
        return new Response(JSON.stringify(catalog), { status: 200 });
      if (url.endsWith("index.json"))
        return new Response(JSON.stringify(providerIndex), { status: 200 });
      return new Response("# doc", { status: 200 });
    }),
  );
}

function renderAt(path: string) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter initialEntries={[path]}>
        <App />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

afterEach(() => vi.restoreAllMocks());

describe("the root route", () => {
  it("orients a visitor and does not list providers", () => {
    stubContract();
    renderAt("/");

    expect(
      screen.getByRole("heading", { name: /nivis registry/i }),
    ).toBeTruthy();
    expect(document.body.textContent).toMatch(
      /derived from the provider binary's own schema/i,
    );
    expect(screen.getByRole("link", { name: /browse providers/i })).toBeTruthy();

    // The catalogue itself is not on the front page.
    expect(screen.queryByText("hashicorp/random")).toBeNull();
    // And the front page fetches nothing.
    expect(fetch).not.toHaveBeenCalled();
  });

  it("explains both compat tiers without claiming the default is proven", () => {
    stubContract();
    renderAt("/");

    expect(screen.getByText(/^compatible by design$/i)).toBeTruthy();
    expect(screen.getByText(/^e2e verified$/i)).toBeTruthy();
    expect(document.body.textContent).toMatch(/does not mean anyone has run it/i);
    expect(document.body.textContent).toMatch(/hand-maintained/i);
  });
});

describe("section routes", () => {
  it("serves the catalogue at its own address", async () => {
    stubContract();
    renderAt("/providers");

    expect(screen.getByRole("heading", { name: /^providers$/i })).toBeTruthy();
    await waitFor(() =>
      expect(screen.getByText("hashicorp/random")).toBeTruthy(),
    );
  });

  it("still renders provider detail routes below it", async () => {
    stubContract();
    renderAt("/providers/hashicorp/random/3.9.0");

    await waitFor(() =>
      expect(screen.getByText("random_password")).toBeTruthy(),
    );
    // The added route depth did not break the relative contract fetch path.
    expect(vi.mocked(fetch).mock.calls[0][0]).toBe(
      "registry/docs/providers/hashicorp/random/3.9.0/index.json",
    );
  });
});

describe("navigation", () => {
  it("is visible from a deep route with the section marked active", async () => {
    stubContract();
    renderAt("/providers/hashicorp/random/3.9.0");

    const link = screen.getByRole("link", { name: "Providers" });
    expect(link).toBeTruthy();
    await waitFor(() => expect(link.getAttribute("aria-current")).toBe("page"));
  });

  it("lists no section that has no page", () => {
    stubContract();
    for (const s of sections) {
      const { unmount } = renderAt(s.to);
      // A section with no route would render an empty <main>.
      expect(document.querySelector("main")?.textContent?.trim()).not.toBe("");
      unmount();
    }
    // Modules is named on the front page, but is not a nav entry yet.
    expect(sections.some((s) => /modules/i.test(s.label))).toBe(false);
  });

  it("returns to the front page through the router, without a document load", async () => {
    stubContract();
    renderAt("/providers/hashicorp/random/3.9.0");

    const before = document.querySelector("main");
    fireEvent.click(screen.getByRole("link", { name: /nivis registry/i }));

    await waitFor(() =>
      expect(screen.getByRole("link", { name: /browse providers/i })).toBeTruthy(),
    );
    // Same <main> element: re-rendered in place, not replaced by a page load.
    expect(document.querySelector("main")).toBe(before);
  });
});
