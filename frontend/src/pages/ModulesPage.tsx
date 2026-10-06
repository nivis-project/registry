import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { fetchModuleCatalog } from "../lib/data";

// ModulesPage lists the catalogued nivis modules. A module's structure is
// derived by evaluating it, so an entry that could NOT be derived is marked:
// the catalogue must not imply every row is machine-checked.
export function ModulesPage() {
  const { data, isLoading, error } = useQuery({
    queryKey: ["module-catalog"],
    queryFn: fetchModuleCatalog,
  });

  return (
    <div>
      <h1 className="text-2xl font-bold text-slate-900">Modules</h1>
      <p className="mt-1 max-w-2xl text-slate-600">
        Composable nivis modules: whole pieces of infrastructure as a single Nix
        import. What each one creates is derived by evaluating the module, not
        copied from its documentation.
      </p>

      {isLoading && <p className="mt-6 text-slate-500">Loading modules…</p>}
      {error && (
        <p className="mt-6 text-red-600">
          Failed to load modules: {String(error)}
        </p>
      )}

      {data && data.length === 0 && (
        <p className="mt-6 text-slate-500">
          No modules are catalogued yet.
        </p>
      )}

      {data && data.length > 0 && (
        <ul className="mt-6 divide-y divide-slate-100 rounded-lg border border-slate-200 bg-white">
          {data.map((m) => (
            <li key={`${m.owner}/${m.name}/${m.version}`}>
              <Link
                to={`/modules/${m.owner}/${m.name}/${m.version}`}
                className="block px-4 py-3 hover:bg-slate-50"
              >
                <div className="flex items-center justify-between gap-4">
                  <span className="font-mono text-sky-700">
                    {m.owner}/{m.name}
                  </span>
                  <span className="flex shrink-0 items-center gap-3 text-sm text-slate-500">
                    <span>{m.version}</span>
                    {!m.derived && (
                      <span className="rounded-full bg-amber-100 px-2 py-0.5 text-xs text-amber-900">
                        structure not derived
                      </span>
                    )}
                  </span>
                </div>
                {m.description && (
                  <p className="mt-1 text-sm text-slate-600">{m.description}</p>
                )}
              </Link>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
