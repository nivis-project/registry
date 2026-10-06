import { Link } from "react-router-dom";

// HomePage orients a first-time visitor before dropping them into a catalogue.
// Deliberately fetches nothing: this is the page that has to render when the
// contract is missing or stale, so it carries no loading or error state, and no
// provider count (a live one needs a fetch, a hardcoded one goes stale).
export function HomePage() {
  return (
    <div>
      <h1 className="text-3xl font-bold text-slate-900">Nivis Registry</h1>
      <p className="mt-2 max-w-2xl text-slate-600">
        OpenTofu-compatible providers, documented as Nix constructors. Every
        reference here is derived from the provider binary's own schema, not
        scraped from its documentation, so it always matches the binary you
        actually run.
      </p>

      <div className="mt-8 grid gap-4 sm:grid-cols-2">
        <Link
          to="/providers"
          className="block rounded-lg border border-slate-200 bg-white p-5 hover:border-sky-300 hover:bg-sky-50"
        >
          <h2 className="font-semibold text-slate-900">Providers</h2>
          <p className="mt-1 text-sm text-slate-600">
            Browse the catalogue. Each provider lists its resources and data
            sources as typed Nix constructors, with a compatibility badge.
          </p>
          <span className="mt-3 inline-block text-sm font-medium text-sky-700">
            Browse providers →
          </span>
        </Link>

        <Link
          to="/modules"
          className="block rounded-lg border border-slate-200 bg-white p-5 hover:border-sky-300 hover:bg-sky-50"
        >
          <h2 className="font-semibold text-slate-900">Modules</h2>
          <p className="mt-1 text-sm text-slate-600">
            Composable nivis modules: whole pieces of infrastructure as a single
            Nix import. What each one creates is derived by evaluating it.
          </p>
          <span className="mt-3 inline-block text-sm font-medium text-sky-700">
            Browse modules →
          </span>
        </Link>
      </div>

      <section className="mt-10 max-w-2xl">
        <h2 className="text-lg font-semibold text-slate-800">
          What the badge means
        </h2>
        <dl className="mt-3 space-y-3 text-sm">
          <div>
            <dt className="font-medium text-slate-900">Compatible by design</dt>
            <dd className="text-slate-600">
              The pipeline downloaded the provider, verified it, and read its
              schema; its protocol version and published architectures are
              recorded. This is the default, and it is a machine-checked claim
              about the provider's interface. It does not mean anyone has run
              it.
            </dd>
          </div>
          <div>
            <dt className="font-medium text-slate-900">E2E verified</dt>
            <dd className="text-slate-600">
              A separate, smaller, hand-maintained set: providers actually
              driven end to end against real infrastructure. A provider is never
              shown as verified without an explicit entry on that list.
            </dd>
          </div>
        </dl>
      </section>
    </div>
  );
}
