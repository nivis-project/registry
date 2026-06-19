import { afterEach, describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ResourcePage } from "./ResourcePage";

// The real nivis gen output for random_password (Nix, never HCL).
const nixDoc = `# \`random_password\`

Nivis Nix constructor for the \`random_password\` resource (provider \`random\`).

## Constructor

\`\`\`nix
{ name, length, keepers ? null, overrides ? {} }
\`\`\`

## Computed outputs

Read these from the resource's \`refAttr\`:

- \`result\`
`;

const index = {
  id: "3.9.0",
  docs: { resources: [], datasources: [], functions: [], guides: [] },
  compat: {
    address: "hashicorp/random",
    tier: "compatible by design",
    schema_extractable: true,
    protocols: ["5.0"],
    architectures: ["linux/amd64", "linux/arm64"],
    e2e: "none",
  },
};

afterEach(() => vi.restoreAllMocks());

function renderResource() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter
        initialEntries={[
          "/providers/hashicorp/random/3.9.0/resources/random_password",
        ]}
      >
        <Routes>
          <Route
            path="/providers/:namespace/:name/:version/resources/:item"
            element={<ResourcePage />}
          />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe("ResourcePage", () => {
  it("renders the Nix constructor document and the compat badge, not HCL", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async (url: string) => {
        if (url.endsWith("index.json")) {
          return new Response(JSON.stringify(index), { status: 200 });
        }
        if (url.endsWith("random_password.md")) {
          return new Response(nixDoc, { status: 200 });
        }
        return new Response("not found", { status: 404 });
      }),
    );

    renderResource();

    // The Nix constructor signature renders.
    await waitFor(() =>
      expect(screen.getByText(/overrides \? \{\}/)).toBeTruthy(),
    );

    // The compat badge surfaces "compatible by design" (never "verified" here).
    expect(screen.getByText(/compatible by design/i)).toBeTruthy();
    expect(screen.queryByText(/e2e verified/i)).toBeNull();

    // No HCL resource block in the rendered output.
    expect(document.body.textContent).not.toMatch(/resource\s+"random_password"/);
  });
});
