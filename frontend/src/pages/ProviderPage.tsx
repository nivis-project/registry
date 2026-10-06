import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Link, useParams } from "react-router-dom";
import { fetchProviderVersion } from "../lib/data";
import { CompatBadge } from "../components/CompatBadge";
import { CodeBlock } from "../components/CodeBlock";
import { ErrorState, SkeletonRows } from "../components/States";
import { providerPage as copy } from "../content/site";
import type { DocItem } from "../lib/contract";
import { useDocumentTitle } from "../lib/useDocumentTitle";

type TabKey = "resources" | "datasources" | "functions";

export function ProviderPage() {
  const { namespace = "", name = "", version = "" } = useParams();
  const ref = { namespace, name, version };
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ["provider", namespace, name, version],
    queryFn: () => fetchProviderVersion(ref),
  });
  const [tab, setTab] = useState<TabKey | null>(null);
  const [filter, setFilter] = useState("");
  useDocumentTitle(`${namespace}/${name}`);

  const tabs = useMemo(() => {
    if (!data) return [] as { key: TabKey; label: string; items: DocItem[] }[];
    return (
      [
        { key: "resources" as const, label: copy.tabs.resources, items: data.docs.resources },
        { key: "datasources" as const, label: copy.tabs.datasources, items: data.docs.datasources },
        { key: "functions" as const, label: copy.tabs.functions, items: data.docs.functions },
      ] satisfies { key: TabKey; label: string; items: DocItem[] }[]
    ).filter((t) => t.items.length > 0);
  }, [data]);

  if (isLoading) return <SkeletonRows rows={8} />;
  if (error || !data)
    return (
      <ErrorState
        what={`registry/docs/providers/${namespace}/${name}/${version}/index.json`}
        error={error}
        onRetry={() => refetch()}
      />
    );

  const active = tab && tabs.some((t) => t.key === tab) ? tab : tabs[0]?.key;
  const items = tabs.find((t) => t.key === active)?.items ?? [];
  const q = filter.trim().toLowerCase();
  const shown = q ? items.filter((i) => i.title.toLowerCase().includes(q)) : items;

  const useSnippet = `lib.mkProvider {\n  source = "registry.opentofu.org/${namespace}/${name}";\n}`;

  return (
    <div>
      <nav aria-label="Breadcrumb" className="text-[14px] text-muted">
        <Link to="/providers" className="hover:text-ink">
          Providers
        </Link>
        <span className="px-1.5">/</span>
        <span className="font-mono">{namespace}</span>
      </nav>

      <div className="mt-2 flex flex-wrap items-center gap-3">
        <h1 className="font-mono text-[clamp(28px,3.4vw,40px)] font-medium text-ink">
          {namespace}/{name}
        </h1>
        <span className="font-mono text-[15px] text-muted">{data.id}</span>
      </div>

      <div className="mt-6 grid gap-6 lg:grid-cols-[1fr_20rem]">
        <div className="min-w-0 order-2 lg:order-1">
          <h2 className="text-[15px] font-semibold text-ink">{copy.useThis}</h2>
          <div className="mt-2">
            <CodeBlock code={useSnippet} lang="nix" />
          </div>

          {tabs.length > 0 && (
            <div className="mt-8">
              <div role="tablist" aria-label="Documentation" className="flex flex-wrap gap-1 border-b border-line">
                {tabs.map((t) => (
                  <button
                    key={t.key}
                    role="tab"
                    id={`tab-${t.key}`}
                    aria-selected={active === t.key}
                    aria-controls={`panel-${t.key}`}
                    onClick={() => setTab(t.key)}
                    className={[
                      "inline-flex h-11 items-center border-b-2 px-3 text-[15px]",
                      active === t.key
                        ? "border-warm font-semibold text-ink"
                        : "border-transparent text-muted hover:text-ink",
                    ].join(" ")}
                  >
                    {t.label}
                    <span className="ml-2 font-mono text-[13px] text-muted">{t.items.length}</span>
                  </button>
                ))}
              </div>

              <div className="mt-4">
                <label className="flex h-11 max-w-sm items-center gap-2 rounded-box border border-line bg-surface px-3">
                  <span className="sr-only">{copy.filterLabel}</span>
                  <input
                    type="search"
                    value={filter}
                    onChange={(e) => setFilter(e.target.value)}
                    placeholder={copy.filterPlaceholder}
                    className="w-full bg-transparent font-mono text-[15px] text-ink outline-none placeholder:text-muted"
                  />
                </label>
              </div>

              <div
                role="tabpanel"
                id={`panel-${active}`}
                aria-labelledby={`tab-${active}`}
                className="mt-4"
              >
                {shown.length === 0 ? (
                  <p className="rounded-card border border-line bg-surface px-4 py-6 text-muted">
                    {copy.noMatch}
                  </p>
                ) : (
                  <ul className="divide-y divide-line overflow-hidden rounded-card border border-line bg-surface">
                    {shown.map((it) => (
                      <li key={it.name}>
                        <Link
                          to={`/providers/${namespace}/${name}/${version}/resources/${it.name}`}
                          className="flex flex-wrap items-baseline gap-x-3 gap-y-1 px-4 py-3 hover:bg-accent-soft"
                        >
                          <span className="font-mono text-[15px] text-accent">{it.title}</span>
                          {it.subcategory && (
                            <span className="rounded-full bg-accent-soft px-2 py-0.5 text-[12px] text-ink">
                              {it.subcategory}
                            </span>
                          )}
                          {it.description && (
                            <span className="text-[14px] text-muted">{it.description}</span>
                          )}
                        </Link>
                      </li>
                    ))}
                  </ul>
                )}
              </div>
            </div>
          )}
        </div>

        <aside className="order-1 lg:order-2">
          <CompatBadge compat={data.compat} />
        </aside>
      </div>
    </div>
  );
}
