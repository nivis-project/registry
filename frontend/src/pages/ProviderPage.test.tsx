import { afterEach, describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ProviderPage } from "./ProviderPage";
import { ProvidersPage } from "./ProvidersPage";

const index = (over: Record<string, unknown> = {}) => ({
  id: "3.9.0",
  docs: {
    resources: [
      { name: "random_password", title: "random_password", description: "a password" },
      { name: "random_pet", title: "random_pet", subcategory: "names" },
    ],
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
  ...over,
});

function stub(body: unknown, ok = true) {
  vi.stubGlobal(
    "fetch",
    vi.fn(async () =>
      ok ? new Response(JSON.stringify(body), { status: 200 }) : new Response("no", { status: 500 }),
    ),
  );
}

function renderProvider() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter initialEntries={["/providers/hashicorp/random/3.9.0"]}>
        <Routes>
          <Route path="/providers/:namespace/:name/:version" element={<ProviderPage />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

afterEach(() => vi.restoreAllMocks());

describe("the provider page", () => {
  it("hides a tab whose list is empty", async () => {
    stub(index());
    renderProvider();

    await waitFor(() => expect(screen.getByRole("tab", { name: /resources/i })).toBeTruthy());
    expect(screen.queryByRole("tab", { name: /data sources/i })).toBeNull();
    expect(screen.queryByRole("tab", { name: /functions/i })).toBeNull();
  });

  it("shows a tab once its list is non-empty", async () => {
    stub(
      index({
        docs: {
          resources: [{ name: "r", title: "r" }],
          datasources: [{ name: "d", title: "d" }],
          functions: [],
          guides: [],
        },
      }),
    );
    renderProvider();
    await waitFor(() => expect(screen.getByRole("tab", { name: /data sources/i })).toBeTruthy());
  });

  it("filters the item list and says when nothing matches", async () => {
    stub(index());
    renderProvider();

    await waitFor(() => expect(screen.getByText("random_password")).toBeTruthy());
    const filter = screen.getByPlaceholderText(/filter by name/i);

    fireEvent.change(filter, { target: { value: "pet" } });
    expect(screen.queryByText("random_password")).toBeNull();
    expect(screen.getByText("random_pet")).toBeTruthy();

    fireEvent.change(filter, { target: { value: "zzz" } });
    expect(screen.getByText(/no item matches/i)).toBeTruthy();
  });

  it("builds the usage snippet from the address, not from sample data", async () => {
    stub(index());
    renderProvider();
    await waitFor(() =>
      expect(document.body.textContent).toMatch(
        /registry\.opentofu\.org\/hashicorp\/random/,
      ),
    );
  });

  it("offers a retry when the contract cannot be fetched", async () => {
    stub(null, false);
    renderProvider();
    await waitFor(() => expect(screen.getByRole("alert")).toBeTruthy());
    expect(screen.getByRole("button", { name: /try again/i })).toBeTruthy();
  });
});

describe("the provider catalogue", () => {
  function renderCatalogue() {
    const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return render(
      <QueryClientProvider client={qc}>
        <MemoryRouter initialEntries={["/providers"]}>
          <ProvidersPage />
        </MemoryRouter>
      </QueryClientProvider>,
    );
  }

  it("filters by address and says when nothing matches", async () => {
    stub([
      { namespace: "hashicorp", name: "random", version: "3.9.0", tier: "compatible by design" },
      { namespace: "ovh", name: "ovh", version: "1.0.0", tier: "compatible by design" },
    ]);
    renderCatalogue();

    await waitFor(() => expect(screen.getByText("random")).toBeTruthy());
    const filter = screen.getByPlaceholderText(/filter by address/i);

    fireEvent.change(filter, { target: { value: "ovh" } });
    expect(screen.queryByText("random")).toBeNull();

    fireEvent.change(filter, { target: { value: "zzz" } });
    expect(screen.getByText(/no provider matches/i)).toBeTruthy();
  });

  it("states an empty catalogue rather than showing nothing", async () => {
    stub([]);
    renderCatalogue();
    await waitFor(() => expect(screen.getByText(/no providers are catalogued/i)).toBeTruthy());
  });
});

describe("owner avatars", () => {
  const entry = (over: Record<string, unknown> = {}) => ({
    namespace: "ovh",
    name: "ovh",
    version: "1.0.0",
    tier: "compatible by design",
    ...over,
  });

  function renderCatalogue() {
    const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return render(
      <QueryClientProvider client={qc}>
        <MemoryRouter initialEntries={["/providers"]}>
          <ProvidersPage />
        </MemoryRouter>
      </QueryClientProvider>,
    );
  }

  it("shows the avatar on a card when the contract has one", async () => {
    stub([entry({ avatar: "avatars/ovh.png" })]);
    renderCatalogue();
    await waitFor(() => expect(screen.getByRole("img", { name: /ovh logo/i })).toBeTruthy());
  });

  it("renders a card without one when the contract omits it", async () => {
    stub([entry()]);
    renderCatalogue();
    await waitFor(() => expect(screen.getByText("ovh")).toBeTruthy());
    expect(screen.queryByRole("img")).toBeNull();
  });

  it("shows the publisher chip only when upstream reported one", async () => {
    stub([entry({ publisher: "partner" }), entry({ name: "other" })]);
    renderCatalogue();
    await waitFor(() => expect(screen.getByText("partner")).toBeTruthy());
    // Two cards, one chip.
    expect(screen.getAllByText("partner")).toHaveLength(1);
  });

  it("offers a filter pill only for a reason the contract actually has", async () => {
    stub([entry({ reason: "europe" }), entry({ name: "other", reason: "popular" })]);
    renderCatalogue();

    await waitFor(() => expect(screen.getByRole("button", { name: /european service/i })).toBeTruthy());
    // "popular" is our bookkeeping, not a facet a reader wants.
    expect(screen.queryByRole("button", { name: /popular/i })).toBeNull();
    // "utility" is offered in principle, but nothing here carries it.
    expect(screen.queryByRole("button", { name: /^utility$/i })).toBeNull();
  });

  it("filters by the selected facet", async () => {
    stub([entry({ reason: "europe" }), entry({ name: "elsewhere", reason: "popular" })]);
    renderCatalogue();

    await waitFor(() => expect(screen.getByText("elsewhere")).toBeTruthy());
    fireEvent.click(screen.getByRole("button", { name: /european service/i }));
    expect(screen.queryByText("elsewhere")).toBeNull();
    expect(screen.getByText("ovh")).toBeTruthy();

    fireEvent.click(screen.getByRole("button", { name: /^all$/i }));
    expect(screen.getByText("elsewhere")).toBeTruthy();
  });

  it("shows the avatar beside the title on a detail page", async () => {
    stub(index({ avatar: "avatars/hashicorp.png" }));
    renderProvider();
    await waitFor(() => expect(screen.getByRole("img", { name: /hashicorp logo/i })).toBeTruthy());
  });
});
