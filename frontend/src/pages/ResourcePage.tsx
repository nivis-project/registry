import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Link, useParams } from "react-router-dom";
import { fetchItemDoc, fetchProviderVersion } from "../lib/data";
import { MarkdownDoc } from "../components/MarkdownDoc";
import { CompatBadge } from "../components/CompatBadge";
import { ErrorState, SkeletonRows } from "../components/States";
import { providerPage, providers as providersCopy } from "../content/site";
import { useDocumentTitle } from "../lib/useDocumentTitle";

// The adapted rendering component: where upstream registry-ui renders scraped
// HCL markdown, this renders the Nix-constructor document the generator emits.
// There is no HCL `resource {}` block anywhere in the output.
export function ResourcePage() {
  const { namespace = "", name = "", version = "", item = "" } = useParams();
  const ref = { namespace, name, version };
  const [filter, setFilter] = useState("");
  useDocumentTitle(item);

  const doc = useQuery({
    queryKey: ["doc", namespace, name, version, item],
    queryFn: () => fetchItemDoc(ref, item),
  });
  const provider = useQuery({
    queryKey: ["provider", namespace, name, version],
    queryFn: () => fetchProviderVersion(ref),
  });

  const siblings = useMemo(() => {
    const d = provider.data;
    if (!d) return [];
    const all = [...d.docs.resources, ...d.docs.datasources, ...d.docs.functions];
    const q = filter.trim().toLowerCase();
    return q ? all.filter((i) => i.title.toLowerCase().includes(q)) : all;
  }, [provider.data, filter]);

  return (
    <div className="grid gap-8 lg:grid-cols-[18rem_1fr]">
      <aside className="min-w-0">
        <Link
          to={`/providers/${namespace}/${name}/${version}`}
          className="font-mono text-[14px] text-accent hover:underline"
        >
          ← {namespace}/{name}
        </Link>

        <label className="mt-3 flex h-11 items-center gap-2 rounded-box border border-line bg-surface px-3">
          <span className="sr-only">{providerPage.filterLabel}</span>
          <input
            type="search"
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
            placeholder={providerPage.filterPlaceholder}
            className="w-full bg-transparent font-mono text-[14px] text-ink outline-none placeholder:text-muted"
          />
        </label>

        {provider.data && (
          <ul className="mt-3 max-h-[60vh] overflow-y-auto rounded-card border border-line bg-surface">
            {siblings.map((s) => (
              <li key={s.name}>
                <Link
                  to={`/providers/${namespace}/${name}/${version}/resources/${s.name}`}
                  aria-current={s.name === item ? "page" : undefined}
                  className={[
                    "block px-3 py-2 font-mono text-[14px]",
                    s.name === item ? "bg-accent-soft text-ink" : "text-muted hover:text-ink",
                  ].join(" ")}
                >
                  {s.title}
                </Link>
              </li>
            ))}
          </ul>
        )}

        {provider.data && (
          <div className="mt-4">
            <CompatBadge compat={provider.data.compat} />
          </div>
        )}
      </aside>

      <div className="min-w-0">
        {doc.isLoading && <SkeletonRows rows={6} />}
        {doc.error && (
          <ErrorState
            what={`registry/docs/providers/${namespace}/${name}/${version}/${item}.md`}
            error={doc.error}
            onRetry={() => doc.refetch()}
          />
        )}
        {doc.data && <MarkdownDoc md={doc.data} />}
        <p className="mt-10 border-t border-line pt-4 text-[14px] text-muted">
          {providersCopy.generatedNote}
        </p>
      </div>
    </div>
  );
}
