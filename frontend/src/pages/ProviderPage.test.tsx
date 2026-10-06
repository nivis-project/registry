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
