import { useQuery } from "@tanstack/react-query";
import { Link, useParams } from "react-router-dom";
import { fetchProviderVersion } from "../lib/data";
import { CompatBadge } from "../components/CompatBadge";
import type { DocItem } from "../lib/contract";

// ProviderPage lists a provider version's resources (datasources/functions when
// nivis gen emits them) and surfaces the compat badge. Each item links to its
// Nix-constructor resource page.
export function ProviderPage() {
  const { namespace = "", name = "", version = "" } = useParams();
  const ref = { namespace, name, version };
  const { data, isLoading, error } = useQuery({
    queryKey: ["provider", namespace, name, version],
    queryFn: () => fetchProviderVersion(ref),
  });

  if (isLoading) return <p className="text-slate-500">Loading…</p>;
  if (error || !data)
    return <p className="text-red-600">Failed to load provider: {String(error)}</p>;

  const section = (title: string, items: DocItem[]) =>
    items.length > 0 && (
      <section className="mt-6">
        <h2 className="text-lg font-semibold text-slate-800">{title}</h2>
        <ul className="mt-2 divide-y divide-slate-100 rounded-lg border border-slate-200 bg-white">
          {items.map((it) => (
            <li key={it.name}>
              <Link
                to={`/providers/${namespace}/${name}/${version}/resources/${it.name}`}
                className="block px-4 py-3 hover:bg-slate-50"
              >
                <span className="font-mono text-sky-700">{it.title}</span>
                {it.description && (
                  <span className="ml-2 text-sm text-slate-500">{it.description}</span>
                )}
              </Link>
            </li>
          ))}
        </ul>
      </section>
    );

  return (
    <div>
      <nav className="text-sm text-slate-500">
        <Link to="/" className="hover:underline">
          Registry
        </Link>{" "}
        / {namespace} /{" "}
        <span className="font-semibold text-slate-700">
          {name} {version}
        </span>
      </nav>
      <h1 className="mt-2 text-2xl font-bold text-slate-900">
        {namespace}/{name}
      </h1>
      <p className="text-slate-500">version {data.id}</p>

      <div className="mt-4 max-w-md">
        <CompatBadge compat={data.compat} />
      </div>

      {section("Resources", data.docs.resources)}
      {section("Data sources", data.docs.datasources)}
      {section("Functions", data.docs.functions)}
    </div>
  );
}
