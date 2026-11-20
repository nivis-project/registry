import { describe, expect, it } from "vitest";
import { QueryClient } from "@tanstack/react-query";
import { headFor, headTags } from "./head";
import { catalogQuery, matchRoute, providerQuery } from "./queries";

const qc = new QueryClient();
const ref = { namespace: "hashicorp", name: "random", version: "3.9.0" };
qc.setQueryData(catalogQuery().queryKey, [{}, {}, {}]);
qc.setQueryData(providerQuery(ref).queryKey, {
  docs: { resources: [{ name: "a" }, { name: "b" }], datasources: [], functions: [], guides: [] },
});

const head = (path: string) => headFor(matchRoute(path), path, qc);

describe("page heads", () => {
  it("canonical URLs are absolute, on registry.nivis.tf, and end in a slash", () => {
    expect(head("/").canonical).toBe("https://registry.nivis.tf/");
    expect(head("/providers").canonical).toBe("https://registry.nivis.tf/providers/");
    expect(head("/providers/hashicorp/random/3.9.0/").canonical).toBe(
      "https://registry.nivis.tf/providers/hashicorp/random/3.9.0/",
    );
  });

  it("a provider page names the provider and counts its resources", () => {
    const h = head("/providers/hashicorp/random/3.9.0");
    expect(h.title).toBe("hashicorp/random · Nivis Registry");
    expect(h.description).toContain("hashicorp/random 3.9.0");
    expect(h.description).toContain("2 resources");
  });

  it("a resource page names the resource and its provider", () => {
    const h = head("/providers/hashicorp/random/3.9.0/resources/random_password");
    expect(h.title).toBe("random_password · Nivis Registry");
    expect(h.description).toContain("random_password from hashicorp/random 3.9.0");
  });

  it("the catalogue counts its providers", () => {
    expect(head("/providers").description).toMatch(/^3 OpenTofu providers/);
  });

  it("an unknown path is not indexed and has no canonical", () => {
    const tags = headTags(head("/nope"));
    expect(tags).toContain('<meta name="robots" content="noindex">');
    expect(tags).not.toContain('rel="canonical"');
  });

  it("every page carries the preview image with its type and size", () => {
    const tags = headTags(head("/providers"));
    expect(tags).toContain('<meta property="og:image" content="https://registry.nivis.tf/social.png">');
    expect(tags).toContain('<meta property="og:image:type" content="image/png">');
    expect(tags).toContain('<meta property="og:image:width" content="1200">');
    expect(tags).toContain('<meta property="og:image:height" content="630">');
    expect(tags).toContain('<link rel="canonical" href="https://registry.nivis.tf/providers/">');
  });

  it("text from the contract is escaped", () => {
    const tags = headTags({ title: 'a "b" <c>', description: "x & y", canonical: "https://registry.nivis.tf/", indexable: true });
    expect(tags).toContain("<title>a &quot;b&quot; &lt;c&gt;</title>");
    expect(tags).toContain('content="x &amp; y"');
  });
});
