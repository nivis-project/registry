import { afterEach, describe, expect, it } from "vitest";
import { QueryClient } from "@tanstack/react-query";
import { setTextFetcher } from "./data";
import {
  catalogQuery,
  docQuery,
  fragmentToPath,
  loadRoute,
  matchRoute,
  moduleCatalogQuery,
  moduleQuery,
  providerQuery,
} from "./queries";

const ref = { namespace: "hashicorp", name: "random", version: "3.9.0" };
const files: Record<string, string> = {
  "/registry/catalog.json": "[]",
  "/registry/modules.json": "[]",
  "/registry/docs/providers/hashicorp/random/3.9.0/index.json": JSON.stringify({ id: "3.9.0" }),
  "/registry/docs/providers/hashicorp/random/3.9.0/random_password.md": "# random_password",
  "/registry/docs/modules/nivis-project/site/1.0.0/index.json": JSON.stringify({ description: "d" }),
};
const asked: string[] = [];

function client() {
  asked.length = 0;
  setTextFetcher(async (url) => {
    asked.push(url);
    if (!(url in files)) throw new Error(`missing ${url}`);
    return files[url];
  });
  return new QueryClient({ defaultOptions: { queries: { retry: false } } });
}

afterEach(() => setTextFetcher(async () => ""));

describe("matchRoute", () => {
  it.each([
    ["/", "home"],
    ["/providers", "providers"],
    ["/providers/", "providers"],
    ["/providers/hashicorp/random/3.9.0", "provider"],
    ["/providers/hashicorp/random/3.9.0/", "provider"],
    ["/providers/hashicorp/random/3.9.0/resources/random_password", "resource"],
    ["/modules", "modules"],
    ["/modules/nivis-project/site/1.0.0", "module"],
    ["/nope", "unknown"],
  ])("%s is a %s route", (path, kind) => {
    expect(matchRoute(path).kind).toBe(kind);
  });
});

describe("loadRoute fills exactly the queries the page reads", () => {
  it("providers", async () => {
    const qc = client();
    await loadRoute(qc, "/providers");
    expect(qc.getQueryData(catalogQuery().queryKey)).toEqual([]);
  });

  it("provider", async () => {
    const qc = client();
    await loadRoute(qc, "/providers/hashicorp/random/3.9.0");
    expect(qc.getQueryData(providerQuery(ref).queryKey)).toEqual({ id: "3.9.0" });
  });

  it("resource: the document and the provider index", async () => {
    const qc = client();
    await loadRoute(qc, "/providers/hashicorp/random/3.9.0/resources/random_password");
    expect(qc.getQueryData(docQuery(ref, "random_password").queryKey)).toBe("# random_password");
    expect(qc.getQueryData(providerQuery(ref).queryKey)).toEqual({ id: "3.9.0" });
  });

  it("modules and a module", async () => {
    const qc = client();
    await loadRoute(qc, "/modules");
    await loadRoute(qc, "/modules/nivis-project/site/1.0.0");
    expect(qc.getQueryData(moduleCatalogQuery().queryKey)).toEqual([]);
    expect(
      qc.getQueryData(moduleQuery({ owner: "nivis-project", name: "site", version: "1.0.0" }).queryKey),
    ).toEqual({ description: "d" });
  });

  it("the front page reads nothing", async () => {
    const qc = client();
    await loadRoute(qc, "/");
    expect(asked).toEqual([]);
  });

  it("a missing document fails the load, so a broken contract cannot prerender", async () => {
    const qc = client();
    await expect(
      loadRoute(qc, "/providers/hashicorp/random/3.9.0/resources/absent"),
    ).rejects.toThrow(/missing/);
  });
});

describe("old fragment links", () => {
  it("move to the path they named", () => {
    expect(fragmentToPath("#/providers/hashicorp/aws/6.0.0")).toBe("/providers/hashicorp/aws/6.0.0");
  });

  it("leave an ordinary anchor alone", () => {
    expect(fragmentToPath("#section")).toBeNull();
    expect(fragmentToPath("")).toBeNull();
  });
});
