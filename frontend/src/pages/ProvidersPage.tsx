import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { fetchCatalog } from "../lib/data";
import { providers as copy, providerFacets } from "../content/site";
import { Avatar } from "../components/Avatar";
import { EmptyState, ErrorState, SkeletonRows } from "../components/States";
import { useDocumentTitle } from "../lib/useDocumentTitle";

// Cards show the address only. catalog.json carries namespace, name, version
// and a compat tier that reads "compatible by design" for every entry, so a
// description, a tier chip or filter pills would all have to be invented.
// catalog-enrichment adds the fields; until then the element is absent rather
// than fabricated.
export function ProvidersPage() {
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ["catalog"],
    queryFn: fetchCatalog,
  });
  const [filter, setFilter] = useState("");
  const [facet, setFacet] = useState<string | null>(null);
  useDocumentTitle(copy.title);

  // Offer a pill only when the contract actually has entries for it.
  const facets = useMemo(
    () => providerFacets.filter((f) => data?.some((p) => p.reason === f.reason)),
    [data],
  );

  const shown = useMemo(() => {
    if (!data) return [];
    const q = filter.trim().toLowerCase();
    return data.filter((p) => {
      if (facet && p.reason !== facet) return false;
      if (q && !`${p.namespace}/${p.name}`.toLowerCase().includes(q)) return false;
      return true;
    });
  }, [data, filter, facet]);

  return (
    <div>
      <div className="flex flex-wrap items-baseline justify-between gap-4">
        <h1 className="text-[clamp(28px,3.4vw,40px)] font-semibold text-ink">
          {copy.title}
          {data && <span className="ml-3 font-mono text-[18px] text-muted">{data.length}</span>}
        </h1>
        {data && data.length > 0 && (
          <label className="flex h-11 items-center gap-2 rounded-box border border-line bg-surface px-3">
            <span className="sr-only">{copy.filterLabel}</span>
            <input
              type="search"
              value={filter}
              onChange={(e) => setFilter(e.target.value)}
              placeholder={copy.filterPlaceholder}
              className="w-56 bg-transparent font-mono text-[15px] text-ink outline-none placeholder:text-muted"
            />
          </label>
        )}
      </div>

      <p className="mt-2 max-w-2xl text-[15px] text-muted">{copy.generatedNote}</p>

      {facets.length > 0 && (
        <div role="group" aria-label={copy.filterLabel} className="mt-4 flex flex-wrap gap-2">
          {[{ reason: null, label: copy.allFilter }, ...facets].map((f) => (
            <button
              key={f.label}
              type="button"
              aria-pressed={facet === f.reason}
              onClick={() => setFacet(f.reason)}
              className={[
                "inline-flex h-11 items-center rounded-full border px-4 text-[14px]",
                facet === f.reason
                  ? "border-accent bg-accent-soft font-medium text-ink"
                  : "border-line bg-surface text-muted hover:text-ink",
              ].join(" ")}
            >
              {f.label}
            </button>
          ))}
        </div>
      )}

      <div className="mt-6">
        {isLoading && <SkeletonRows rows={8} />}
        {error && <ErrorState what="registry/catalog.json" error={error} onRetry={() => refetch()} />}
        {data && data.length === 0 && <EmptyState>{copy.empty}</EmptyState>}
        {data && data.length > 0 && shown.length === 0 && <EmptyState>{copy.noMatch}</EmptyState>}

        {shown.length > 0 && (
          <ul className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {shown.map((p) => (
              <li key={`${p.namespace}/${p.name}/${p.version}`}>
                <Link
                  to={`/providers/${p.namespace}/${p.name}/${p.version}`}
                  className="flex h-full flex-col justify-between rounded-card border border-line bg-surface p-4 hover:border-accent"
                >
                  <span className="flex items-center gap-2.5">
                    <Avatar src={p.avatar} owner={p.namespace} size={28} />
                    <span className="min-w-0 font-mono text-[15px]">
                      <span className="text-muted">{p.namespace}/</span>
                      <span className="text-ink">{p.name}</span>
                    </span>
                  </span>
                  <span className="mt-3 flex flex-wrap items-center gap-2">
                    <span className="rounded-full bg-accent-soft px-2.5 py-0.5 text-[13px] text-ink">
                      {p.tier}
                    </span>
                    {p.publisher && (
                      <span className="rounded-full border border-line px-2.5 py-0.5 text-[13px] text-muted">
                        {p.publisher}
                      </span>
                    )}
                    <span className="font-mono text-[13px] text-muted">{p.version}</span>
                  </span>
                </Link>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
}
