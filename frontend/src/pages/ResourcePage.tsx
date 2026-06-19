import { useQuery } from "@tanstack/react-query";
import { Link, useParams } from "react-router-dom";
import { fetchItemDoc, fetchProviderVersion } from "../lib/data";
import { MarkdownDoc } from "../components/MarkdownDoc";
import { CompatBadge } from "../components/CompatBadge";

// ResourcePage is the ADAPTED rendering component (the only Nivis-specific
// change to registry-ui's data-agnostic SPA). Where upstream renders scraped HCL
// markdown, this renders the Nix-constructor document our generator emits — the
// schema-derived Nivis reference. There is no HCL `resource {}` block.
export function ResourcePage() {
  const {
    namespace = "",
    name = "",
    version = "",
    item = "",
  } = useParams();
  const ref = { namespace, name, version };

  const doc = useQuery({
    queryKey: ["doc", namespace, name, version, item],
    queryFn: () => fetchItemDoc(ref, item),
  });
  const provider = useQuery({
    queryKey: ["provider", namespace, name, version],
    queryFn: () => fetchProviderVersion(ref),
  });

  return (
    <div>
      <nav className="text-sm text-slate-500">
        <Link to="/" className="hover:underline">
          Registry
        </Link>{" "}
        /{" "}
        <Link
          to={`/providers/${namespace}/${name}/${version}`}
          className="hover:underline"
        >
          {namespace}/{name} {version}
        </Link>{" "}
        / <span className="font-mono text-slate-700">{item}</span>
      </nav>

      <div className="mt-4 grid gap-6 lg:grid-cols-[1fr_20rem]">
        <article className="rounded-lg border border-slate-200 bg-white p-6">
          {doc.isLoading && <p className="text-slate-500">Loading…</p>}
          {doc.error && (
            <p className="text-red-600">Failed to load document: {String(doc.error)}</p>
          )}
          {doc.data && <MarkdownDoc md={doc.data} />}
        </article>

        <aside className="space-y-4">
          {provider.data && <CompatBadge compat={provider.data.compat} />}
          <div className="rounded-lg border border-slate-200 bg-slate-50 p-4 text-sm text-slate-600">
            This reference is generated from the provider binary's own schema via
            <code className="mx-1 rounded bg-slate-200 px-1">nivis gen</code>, so it
            always matches the real provider. It is a Nix constructor — not scraped
            HCL.
          </div>
        </aside>
      </div>
    </div>
  );
}
