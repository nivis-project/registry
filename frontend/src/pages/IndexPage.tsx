import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { fetchCatalog } from "../lib/data";

// IndexPage lists the providers present in the static contract. v1 has no live
// search backend (that is milestone 06); the catalog.json generated alongside
// the contract is enough to browse and link into each provider.
export function IndexPage() {
  const { data, isLoading, error } = useQuery({
    queryKey: ["catalog"],
    queryFn: fetchCatalog,
  });

  return (
    <div>
      <h1 className="text-2xl font-bold text-slate-900">Nivis Registry</h1>
      <p className="mt-1 text-slate-600">
        OpenTofu-compatible providers with Nix-native documentation. Every
        provider is compatible by design; references are derived from each
        provider's own schema.
      </p>

      {isLoading && <p className="mt-6 text-slate-500">Loading catalog…</p>}
      {error && <p className="mt-6 text-red-600">Failed to load catalog: {String(error)}</p>}

      <ul className="mt-6 divide-y divide-slate-100 rounded-lg border border-slate-200 bg-white">
        {data?.map((p) => (
          <li key={`${p.namespace}/${p.name}/${p.version}`}>
            <Link
              to={`/providers/${p.namespace}/${p.name}/${p.version}`}
              className="flex items-center justify-between px-4 py-3 hover:bg-slate-50"
            >
              <span className="font-mono text-sky-700">
                {p.namespace}/{p.name}
              </span>
              <span className="flex items-center gap-3 text-sm text-slate-500">
                <span>{p.version}</span>
                <span className="rounded-full bg-sky-100 px-2 py-0.5 text-xs text-sky-800">
                  {p.tier}
                </span>
              </span>
            </Link>
          </li>
        ))}
      </ul>
    </div>
  );
}
