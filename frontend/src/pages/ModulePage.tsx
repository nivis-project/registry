import { useQuery } from "@tanstack/react-query";
import { Link, useParams } from "react-router-dom";
import { fetchModuleVersion } from "../lib/data";
import { MarkdownDoc } from "../components/MarkdownDoc";
import { ErrorState, SkeletonRows } from "../components/States";
import { modulePage as copy } from "../content/site";
import { useDocumentTitle } from "../lib/useDocumentTitle";
import type { ModuleCoord } from "../lib/contract";

// providerHref turns a resolved "<ns>/<name>/<version>" reference plus a
// resource type into the provider resource route.
function providerHref(docs: string, type: string): string {
  return `/providers/${docs}/resources/${type}`;
}

// CoordRow renders one resource or data source. A coordinate the extractor
// could not resolve still appears — dropping it would misrepresent what the
// module creates — but carries no link, because there is no page to link to.
function CoordRow({ c }: { c: ModuleCoord }) {
  return (
    <li className="flex flex-wrap items-baseline justify-between gap-2 px-4 py-2">
      <span className="font-mono text-sm">
        {c.docs ? (
          <Link to={providerHref(c.docs, c.type)} className="text-accent hover:underline">
            {c.type}
          </Link>
        ) : (
          <span className="text-ink">{c.type}</span>
        )}
        <span className="text-muted"> · {c.name}</span>
      </span>
      <span className="text-[12px] text-muted">
        {c.docs ? c.docs : copy.uncatalogued}
      </span>
    </li>
  );
}

function CoordList({ title, items }: { title: string; items: ModuleCoord[] }) {
  if (items.length === 0) return null;
  return (
    <section className="mt-6">
      <h2 className="text-[20px] font-semibold text-ink">{title}</h2>
      <ul className="mt-2 divide-y divide-line overflow-hidden rounded-card border border-line bg-surface">
        {items.map((c) => (
          <CoordRow key={`${c.provider}.${c.type}.${c.name}`} c={c} />
        ))}
      </ul>
    </section>
  );
}

export function ModulePage() {
  const { owner = "", name = "", version = "" } = useParams();
  const ref = { owner, name, version };
  useDocumentTitle(`${owner}/${name}`);
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ["module", owner, name, version],
    queryFn: () => fetchModuleVersion(ref),
  });

  if (isLoading) return <SkeletonRows rows={6} />;
  if (error || !data)
    return (
      <ErrorState
        what={`registry/docs/modules/${owner}/${name}/${version}/index.json`}
        error={error}
        onRetry={() => refetch()}
      />
    );

  const required = data.cfg.filter((k) => k.required);
  const optional = data.cfg.filter((k) => !k.required);
  const { record } = data;

  return (
    <div>
      <p className="text-[14px] text-muted">
        <Link to="/modules" className="hover:text-ink">
          Modules
        </Link>
      </p>
      <h1 className="mt-1 font-mono text-[clamp(28px,3.4vw,40px)] font-medium text-ink">
        {data.owner}/{data.name}
      </h1>
      <p className="mt-1 text-[14px] text-muted">
        {data.tag ?? data.id}
        {data.rev && <span className="text-muted"> · {data.rev.slice(0, 7)}</span>}
      </p>
      {data.description && (
        <p className="mt-2 max-w-2xl text-muted">{data.description}</p>
      )}

      {!record.structure_extractable ? (
        <section className="mt-6 rounded-card border border-warning-line bg-warning-bg p-4">
          <h2 className="font-semibold text-warning-ink">{copy.notDerivedTitle}</h2>
          <p className="mt-1 text-[14px] text-warning-ink">
            This module could not be evaluated, so the registry cannot say what it
            creates. This is not a claim that it creates nothing.
          </p>
          {record.failure_reason && (
            <pre className="mt-2 overflow-x-auto rounded-box bg-code-bg p-3 font-mono text-[13px] text-code-ink">
              {record.failure_reason}
            </pre>
          )}
        </section>
      ) : (
        <>
          <CoordList title="Creates" items={data.resources} />
          <CoordList title="Reads" items={data.data_sources} />

          {data.outputs.length > 0 && (
            <section className="mt-6">
              <h2 className="text-[20px] font-semibold text-ink">{copy.outputs}</h2>
              <ul className="mt-2 flex flex-wrap gap-2">
                {data.outputs.map((o) => (
                  <li
                    key={o}
                    className="rounded-full border border-line bg-surface px-2.5 py-1 font-mono text-[14px] text-ink"
                  >
                    {o}
                  </li>
                ))}
              </ul>
            </section>
          )}

          {data.composition && data.composition.length > 0 && (
            <section className="mt-6">
              <h2 className="text-[20px] font-semibold text-ink">{copy.exposes}</h2>
              <p className="mt-1 text-[14px] text-muted">{copy.exposesBody}</p>
              <ul className="mt-2 flex flex-wrap gap-2">
                {data.composition.map((k) => (
                  <li
                    key={k}
                    className="rounded-full border border-line bg-surface px-2.5 py-1 font-mono text-[14px] text-ink"
                  >
                    {k}
                  </li>
                ))}
              </ul>
            </section>
          )}

          <section className="mt-6">
            <h2 className="text-[20px] font-semibold text-ink">{copy.configuration}</h2>
            {record.cfg_source === "scanned" && (
              <p className="mt-1 text-[14px] text-muted">
                Inferred by reading the module's source. A nivis module does not
                declare its configuration, so this list is a best effort, not a
                declaration.
              </p>
            )}
            <div className="mt-2 grid gap-4 sm:grid-cols-2">
              <div>
                <h3 className="text-[14px] font-medium text-ink">{copy.required}</h3>
                <ul className="mt-1 space-y-1">
                  {required.length === 0 && (
                    <li className="text-[14px] text-muted">{copy.none}</li>
                  )}
                  {required.map((k) => (
                    <li key={k.path} className="font-mono text-[14px] text-ink">
                      {k.path}
                    </li>
                  ))}
                </ul>
              </div>
              <div>
                <h3 className="text-[14px] font-medium text-muted">{copy.optional}</h3>
                <ul className="mt-1 space-y-1">
                  {optional.length === 0 && (
                    <li className="text-[14px] text-muted">{copy.none}</li>
                  )}
                  {optional.map((k) => (
                    <li key={k.path} className="font-mono text-[14px] text-muted">
                      {k.path}
                    </li>
                  ))}
                </ul>
              </div>
            </div>
          </section>

          <p className="mt-6 text-[13px] text-muted">
            {record.providers_resolved} of {record.providers_total} resource types
            link to a provider catalogued here.
          </p>
        </>
      )}

      {data.readme && (
        <section className="mt-10 border-t border-line pt-6">
          <h2 className="text-[20px] font-semibold text-ink">{copy.readmeTitle}</h2>
          <p className="mt-1 text-[14px] text-muted">
            Written by the module's authors, not derived from it.
          </p>
          <div className="mt-3">
            <MarkdownDoc md={data.readme} />
          </div>
        </section>
      )}
    </div>
  );
}
