import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { fetchModuleCatalog } from "../lib/data";
import { modules as copy } from "../content/site";
import { EmptyState, ErrorState, SkeletonRows } from "../components/States";
import { Avatar } from "../components/Avatar";
import { useDocumentTitle } from "../lib/useDocumentTitle";

// A module's structure is derived by evaluating it, so an entry that could NOT
// be derived is marked: the catalogue must not imply every row is
// machine-checked.
export function ModulesPage() {
  useDocumentTitle(copy.title);
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ["module-catalog"],
    queryFn: fetchModuleCatalog,
  });

  return (
    <div>
      <h1 className="text-[clamp(28px,3.4vw,40px)] font-semibold text-ink">
        {copy.title}
        {data && <span className="ml-3 font-mono text-[18px] text-muted">{data.length}</span>}
      </h1>
      <p className="mt-2 max-w-2xl text-[15px] text-muted">{copy.lead}</p>

      <div className="mt-6">
        {isLoading && <SkeletonRows rows={4} />}
        {error && <ErrorState what="registry/modules.json" error={error} onRetry={() => refetch()} />}
        {data && data.length === 0 && <EmptyState>{copy.empty}</EmptyState>}

        {data && data.length > 0 && (
          <ul className="divide-y divide-line overflow-hidden rounded-card border border-line bg-surface">
            {data.map((m) => (
              <li key={`${m.owner}/${m.name}/${m.version}`}>
                <Link
                  to={`/modules/${m.owner}/${m.name}/${m.version}`}
                  className="block px-4 py-4 hover:bg-accent-soft"
                >
                  <div className="flex flex-wrap items-center justify-between gap-3">
                    <span className="flex min-w-0 items-center gap-2.5">
                      <Avatar src={m.avatar} owner={m.owner} size={28} />
                      <span className="font-mono text-[15px]">
                        <span className="text-muted">{m.owner}/</span>
                        <span className="text-accent">{m.name}</span>
                      </span>
                    </span>
                    <span className="flex shrink-0 items-center gap-3">
                      {!m.derived && (
                        <span className="rounded-full border border-warning-line bg-warning-bg px-2.5 py-0.5 text-[12px] text-warning-ink">
                          {copy.notDerivedChip}
                        </span>
                      )}
                      <span className="font-mono text-[14px] text-muted">{m.version}</span>
                    </span>
                  </div>
                  {m.description && (
                    <p className="mt-1 text-[14px] text-muted">{m.description}</p>
                  )}
                </Link>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
}
