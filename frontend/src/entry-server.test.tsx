// @vitest-environment node
import { describe, expect, it } from "vitest";
import { render } from "./entry-server";

const files: Record<string, string> = {
  "/registry/docs/providers/hashicorp/random/3.9.0/index.json": JSON.stringify({
    id: "3.9.0",
    docs: {
      resources: [
        { name: "random_password", title: "random_password", description: "a password" },
        { name: "random_string", title: "random_string", description: "a string" },
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
  }),
  "/registry/docs/providers/hashicorp/random/3.9.0/random_password.md":
    "# `random_password`\n\n## Constructor\n\n```nix\n{ name, length, overrides ? {} }\n```\n",
};
files["/registry/docs/providers/hashicorp/http/3.6.3/index.json"] = JSON.stringify({
  id: "3.6.3",
  docs: { resources: null, datasources: [{ name: "http", title: "http" }], functions: null, guides: null },
  compat: {
    address: "hashicorp/http",
    tier: "compatible by design",
    schema_extractable: true,
    protocols: ["5.0"],
    architectures: ["linux/amd64"],
    e2e: "none",
  },
});

const fetcher = async (url: string) => {
  if (!(url in files)) throw new Error(`missing ${url}`);
  return files[url];
};

describe("server render", () => {
  it("a resource page carries its constructor without any script", async () => {
    const page = await render("/providers/hashicorp/random/3.9.0/resources/random_password", fetcher);
    expect(page.html).toContain("random_password");
    expect(page.html).toContain("{ name, length, overrides ? {} }");
    expect(page.headTags).toContain("<title>random_password · Nivis Registry</title>");
  });

  it("leaves the long sibling list to the browser", async () => {
    const page = await render("/providers/hashicorp/random/3.9.0/resources/random_password", fetcher);
    expect(page.html).not.toContain("/resources/random_string");
  });

  it("a provider page lists its resources", async () => {
    const page = await render("/providers/hashicorp/random/3.9.0", fetcher);
    expect(page.html).toContain("/providers/hashicorp/random/3.9.0/resources/random_string");
  });

  it("the header links Home, Docs and Star on GitHub", async () => {
    const page = await render("/", fetcher);
    expect(page.html).toContain('href="https://nivis.tf"');
    expect(page.html).toContain('href="https://docs.nivis.tf/"');
    expect(page.html).toMatch(/href="https:\/\/github.com\/nivis-project\/nivis"[^>]*>.*Star on GitHub/s);
  });

  it("the footer links GitHub, Issues, Discussions and Demos", async () => {
    const page = await render("/", fetcher);
    for (const href of [
      "https://github.com/nivis-project/nivis/issues",
      "https://github.com/nivis-project/nivis/discussions",
      "https://github.com/nivis-project/nivis-demos",
    ]) {
      expect(page.html).toContain(`href="${href}"`);
    }
  });

  it("a provider whose contract has null item lists still renders", async () => {
    const page = await render("/providers/hashicorp/http/3.6.3", fetcher);
    expect(page.html).toContain("/providers/hashicorp/http/3.6.3/resources/http");
    expect(page.head.description).toContain("0 resources");
  });

  it("a missing document fails the render instead of publishing an error page", async () => {
    await expect(
      render("/providers/hashicorp/random/3.9.0/resources/absent", fetcher),
    ).rejects.toThrow(/missing/);
  });
});
