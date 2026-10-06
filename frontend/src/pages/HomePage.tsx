import { Link } from "react-router-dom";
import { home } from "../content/site";
import { useDocumentTitle } from "../lib/useDocumentTitle";

// Fetches nothing, by requirement: this page has to render when the contract is
// missing or stale, so it carries no counts and no loading or error state.
export function HomePage() {
  useDocumentTitle();
  return (
    <div className="-mx-6 -mt-8">
      <section className="bg-band-bg px-6 py-14 text-band-ink">
        <div className="mx-auto max-w-shell">
          <h1 className="max-w-3xl text-[clamp(34px,4.4vw,54px)] font-semibold leading-[1.1]">
            {home.title}
          </h1>
          <p className="mt-4 max-w-2xl text-band-muted">{home.lead}</p>
        </div>
      </section>

      <div className="mx-auto max-w-shell px-6 py-10">
        <div className="grid gap-4 sm:grid-cols-2">
          <EntryCard to="/providers" title={home.providers.title} body={home.providers.body} cta={home.providers.cta} />
          <EntryCard to="/modules" title={home.modules.title} body={home.modules.body} cta={home.modules.cta} />
        </div>

        <section className="mt-12 max-w-2xl">
          <h2 className="text-[26px] font-semibold text-ink">{home.badge.title}</h2>
          <dl className="mt-4 space-y-4 text-[15px]">
            {[home.badge.byDesign, home.badge.verified].map((t) => (
              <div key={t.term}>
                <dt className="font-medium text-ink">{t.term}</dt>
                <dd className="mt-0.5 text-muted">{t.body}</dd>
              </div>
            ))}
          </dl>
        </section>
      </div>
    </div>
  );
}

function EntryCard({ to, title, body, cta }: { to: string; title: string; body: string; cta: string }) {
  return (
    <Link
      to={to}
      className="block rounded-card border border-line bg-surface p-6 hover:border-accent"
    >
      <h2 className="text-[20px] font-semibold text-ink">{title}</h2>
      <p className="mt-1.5 text-[15px] text-muted">{body}</p>
      <span className="mt-4 inline-block text-[15px] font-medium text-accent">{cta} →</span>
    </Link>
  );
}
