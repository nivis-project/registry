import { afterEach, describe, expect, it, vi } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { App } from "../App";
import type { ModuleVersion } from "../lib/contract";

const catalog = [
  {
    owner: "wearetechnative",
    name: "nivis-aws-form-action",
    version: "0.1.0",
    description: "altcha-protected HTML-form backend",
    derived: true,
  },
  {
    owner: "wearetechnative",
    name: "broken-module",
    version: "0.2.0",
    description: "did not evaluate",
    derived: false,
  },
];

const derived: ModuleVersion = {
  id: "0.1.0",
  owner: "wearetechnative",
  name: "nivis-aws-form-action",
  tag: "v0.1.0",
  rev: "8d542f21e91c0a600840a3fa8faa3444f61fd63c",
  description: "altcha-protected HTML-form backend",
  resources: [
    { provider: "aws", type: "aws_lambda_function", name: "altcha", docs: "hashicorp/aws/6.67.0" },
    { provider: "weird", type: "weird_thing", name: "x" },
  ],
  data_sources: [
    { provider: "aws", type: "aws_ssm_parameter", name: "tok", docs: "hashicorp/aws/6.67.0" },
  ],
  outputs: ["api_endpoint", "submit_url"],
  composition: ["apiEndpointRef"],
  cfg: [
    { path: "baseName", required: true },
    { path: "ses.from", required: true },
    { path: "memorySize", required: false },
  ],
  record: {
    structure_extractable: true,
    cfg_source: "scanned",
    providers_resolved: 2,
    providers_total: 3,
  },
  readme: "# form action\n\nHuman written prose.",
};

const notDerived: ModuleVersion = {
  id: "0.2.0",
  owner: "wearetechnative",
  name: "broken-module",
  resources: [],
  data_sources: [],
  outputs: [],
  cfg: [],
  record: {
    structure_extractable: false,
    cfg_source: "scanned",
    providers_resolved: 0,
    providers_total: 0,
    failure_reason: "error: module read cfg while probing its structure",
  },
};

function stub(modules: Record<string, ModuleVersion>, cat = catalog) {
  vi.stubGlobal(
    "fetch",
    vi.fn(async (url: string) => {
      if (url.endsWith("registry/modules.json"))
        return new Response(JSON.stringify(cat), { status: 200 });
      for (const [key, body] of Object.entries(modules)) {
        if (url.includes(key))
          return new Response(JSON.stringify(body), { status: 200 });
      }
      return new Response("not found", { status: 404 });
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

describe("the module catalogue", () => {
  it("lists what the contract holds", async () => {
    stub({});
    renderAt("/modules");

    const link = await waitFor(() =>
      screen.getByRole("link", { name: /nivis-aws-form-action/ }),
    );
    expect(link.getAttribute("href")).toContain(
      "/modules/wearetechnative/nivis-aws-form-action/0.1.0",
    );
  });

  it("marks an entry whose structure was not derived", async () => {
    stub({});
    renderAt("/modules");

    await waitFor(() => expect(screen.getByText(/structure not derived/i)).toBeTruthy());
    // The derived one carries no such marker.
    const derivedRow = screen.getByRole("link", { name: /nivis-aws-form-action/ });
    expect(within(derivedRow).queryByText(/structure not derived/i)).toBeNull();
  });

  it("states plainly when nothing is catalogued", async () => {
    stub({}, []);
    renderAt("/modules");

    await waitFor(() =>
      expect(screen.getByText(/no modules are catalogued yet/i)).toBeTruthy(),
    );
  });
});

describe("a module page", () => {
  it("leads with the derived structure and keeps the README separate", async () => {
    stub({ "nivis-aws-form-action": derived });
    renderAt("/modules/wearetechnative/nivis-aws-form-action/0.1.0");

    await waitFor(() => expect(screen.getByText("Creates")).toBeTruthy());
    expect(screen.getByText("Reads")).toBeTruthy();
    expect(screen.getByText("api_endpoint")).toBeTruthy();
    expect(screen.getByText("apiEndpointRef")).toBeTruthy();

    // The README is present but identified as human-written, not derived.
    expect(screen.getByText(/from the module's readme/i)).toBeTruthy();
    expect(screen.getByText(/written by the module's authors/i)).toBeTruthy();
  });

  it("links a resolved resource type to its provider page", async () => {
    stub({ "nivis-aws-form-action": derived });
    renderAt("/modules/wearetechnative/nivis-aws-form-action/0.1.0");

    const link = await waitFor(() =>
      screen.getByRole("link", { name: "aws_lambda_function" }),
    );
    expect(link.getAttribute("href")).toContain(
      "/providers/hashicorp/aws/6.67.0/resources/aws_lambda_function",
    );
  });

  it("shows an uncatalogued resource type without a link", async () => {
    stub({ "nivis-aws-form-action": derived });
    renderAt("/modules/wearetechnative/nivis-aws-form-action/0.1.0");

    await waitFor(() => expect(screen.getByText("weird_thing")).toBeTruthy());
    // Present, but not a link, and the reason is stated.
    expect(screen.queryByRole("link", { name: "weird_thing" })).toBeNull();
    expect(screen.getByText(/provider not catalogued here/i)).toBeTruthy();
  });

  it("separates required from optional configuration", async () => {
    stub({ "nivis-aws-form-action": derived });
    renderAt("/modules/wearetechnative/nivis-aws-form-action/0.1.0");

    const requiredHeading = await waitFor(() =>
      screen.getByRole("heading", { name: /^required$/i }),
    );
    const requiredGroup = requiredHeading.parentElement!;
    expect(within(requiredGroup).getByText("baseName")).toBeTruthy();
    expect(within(requiredGroup).getByText("ses.from")).toBeTruthy();
    expect(within(requiredGroup).queryByText("memorySize")).toBeNull();

    const optionalGroup = screen.getByRole("heading", { name: /^optional$/i })
      .parentElement!;
    expect(within(optionalGroup).getByText("memorySize")).toBeTruthy();
  });

  it("says the configuration surface is inferred", async () => {
    stub({ "nivis-aws-form-action": derived });
    renderAt("/modules/wearetechnative/nivis-aws-form-action/0.1.0");

    await waitFor(() =>
      expect(screen.getByText(/inferred by reading the module's source/i)).toBeTruthy(),
    );
    expect(document.body.textContent).toMatch(
      /2 of 3 resource types link to a provider/i,
    );
  });

  it("states that a module did not evaluate instead of showing an empty structure", async () => {
    stub({ "broken-module": notDerived });
    renderAt("/modules/wearetechnative/broken-module/0.2.0");

    await waitFor(() =>
      expect(screen.getByText(/structure not derived/i)).toBeTruthy(),
    );
    expect(document.body.textContent).toMatch(/not a claim that it creates nothing/i);
    expect(screen.getByText(/module read cfg while probing/i)).toBeTruthy();
    // No empty "Creates" section implying the module builds nothing.
    expect(screen.queryByText("Creates")).toBeNull();
  });
});
