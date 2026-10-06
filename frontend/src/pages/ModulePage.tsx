import { useQuery } from "@tanstack/react-query";
import { Link, useParams } from "react-router-dom";
import { fetchModuleVersion } from "../lib/data";
import { MarkdownDoc } from "../components/MarkdownDoc";
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
          <Link to={providerHref(c.docs, c.type)} className="text-sky-700 hover:underline">
            {c.type}
          </Link>
        ) : (
          <span className="text-slate-800">{c.type}</span>
        )}
        <span className="text-slate-400"> · {c.name}</span>
      </span>
      <span className="text-xs text-slate-500">
        {c.docs ? c.docs : "provider not catalogued here"}
      </span>
    </li>
  );
}

function CoordList({ title, items }: { title: string; items: ModuleCoord[] }) {
  if (items.length === 0) return null;
  return (
    <section className="mt-6">
      <h2 className="text-lg font-semibold text-slate-800">{title}</h2>
      <ul className="mt-2 divide-y divide-slate-100 rounded-lg border border-slate-200 bg-white">
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
  const { data, isLoading, error } = useQuery({
    queryKey: ["module", owner, name, version],
    queryFn: () => fetchModuleVersion(ref),
  });

  if (isLoading) return <p className="text-slate-500">Loading…</p>;
  if (error || !data)
    return <p className="text-red-600">Failed to load module: {String(error)}</p>;

  const required = data.cfg.filter((k) => k.required);
  const optional = data.cfg.filter((k) => !k.required);
  const { record } = data;

  return (
    <div>
      <p className="text-sm text-slate-500">
        <Link to="/modules" className="hover:underline">
          Modules
        </Link>
      </p>
      <h1 className="mt-1 font-mono text-2xl font-bold text-slate-900">
        {data.owner}/{data.name}
      </h1>
      <p className="mt-1 text-sm text-slate-500">
        {data.tag ?? data.id}
        {data.rev && <span className="text-slate-400"> · {data.rev.slice(0, 7)}</span>}
      </p>
      {data.description && (
        <p className="mt-2 max-w-2xl text-slate-600">{data.description}</p>
      )}

      {!record.structure_extractable ? (
        <section className="mt-6 rounded-lg border border-amber-200 bg-amber-50 p-4">
          <h2 className="font-semibold text-amber-900">Structure not derived</h2>
          <p className="mt-1 text-sm text-amber-900">
            This module could not be evaluated, so the registry cannot say what it
            creates. This is not a claim that it creates nothing.
          </p>
          {record.failure_reason && (
            <pre className="mt-2 overflow-x-auto rounded bg-amber-100 p-2 text-xs text-amber-900">
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
              <h2 className="text-lg font-semibold text-slate-800">Outputs</h2>
              <ul className="mt-2 flex flex-wrap gap-2">
                {data.outputs.map((o) => (
                  <li
                    key={o}
                    className="rounded-md bg-white px-2 py-1 font-mono text-sm text-slate-700 ring-1 ring-slate-200"
                  >
                    {o}
                  </li>
                ))}
              </ul>
            </section>
          )}

          {data.composition && data.composition.length > 0 && (
            <section className="mt-6">
              <h2 className="text-lg font-semibold text-slate-800">
                Also exposes
              </h2>
              <p className="mt-1 text-sm text-slate-600">
                Further values this module returns, which another module can
                consume.
              </p>
              <ul className="mt-2 flex flex-wrap gap-2">
                {data.composition.map((k) => (
                  <li
                    key={k}
                    className="rounded-md bg-white px-2 py-1 font-mono text-sm text-slate-700 ring-1 ring-slate-200"
                  >
                    {k}
                  </li>
                ))}
              </ul>
            </section>
          )}

          <section className="mt-6">
            <h2 className="text-lg font-semibold text-slate-800">Configuration</h2>
            {record.cfg_source === "scanned" && (
              <p className="mt-1 text-sm text-slate-500">
                Inferred by reading the module's source. A nivis module does not
                declare its configuration, so this list is a best effort, not a
                declaration.
              </p>
            )}
            <div className="mt-2 grid gap-4 sm:grid-cols-2">
              <div>
                <h3 className="text-sm font-medium text-slate-900">Required</h3>
                <ul className="mt-1 space-y-1">
                  {required.length === 0 && (
                    <li className="text-sm text-slate-500">None.</li>
                  )}
                  {required.map((k) => (
                    <li key={k.path} className="font-mono text-sm text-slate-700">
                      {k.path}
                    </li>
                  ))}
                </ul>
              </div>
              <div>
                <h3 className="text-sm font-medium text-slate-500">Optional</h3>
                <ul className="mt-1 space-y-1">
                  {optional.length === 0 && (
                    <li className="text-sm text-slate-500">None.</li>
                  )}
                  {optional.map((k) => (
                    <li key={k.path} className="font-mono text-sm text-slate-500">
                      {k.path}
                    </li>
                  ))}
                </ul>
              </div>
            </div>
          </section>

          <p className="mt-6 text-xs text-slate-500">
            {record.providers_resolved} of {record.providers_total} resource types
            link to a provider catalogued here.
          </p>
        </>
      )}

      {data.readme && (
        <section className="mt-10 border-t border-slate-200 pt-6">
          <h2 className="text-lg font-semibold text-slate-800">
            From the module's README
          </h2>
          <p className="mt-1 text-sm text-slate-500">
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
