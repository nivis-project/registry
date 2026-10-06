import type { CompatRecord } from "../lib/contract";
import { providerPage } from "../content/site";

// The compatibility panel. It never shows "verified" unless the contract says
// so, and every axis falls back to a dash rather than to an optimistic guess.
const DASH = "-";

export function CompatBadge({ compat }: { compat: CompatRecord }) {
  const verified = compat.tier === "e2e verified" || compat.e2e === "verified";
  const rows: [string, string][] = [
    ["Schema", compat.schema_extractable ? "extracted" : DASH],
    ["Protocol", compat.protocols.join(", ") || DASH],
    ["Systems", compat.architectures.join(", ") || DASH],
    ["End to end", compat.e2e || DASH],
  ];

  return (
    <section
      aria-label={providerPage.compatTitle}
      className="rounded-card border border-line bg-surface p-5"
    >
      <h2 className="text-[15px] font-semibold text-ink">{providerPage.compatTitle}</h2>
      <p className="mt-2">
        <span className="inline-flex items-center rounded-full bg-accent-soft px-3 py-1 text-[14px] font-medium text-ink">
          {verified ? "e2e verified" : "compatible by design"}
        </span>
      </p>
      <dl className="mt-4 grid grid-cols-[7.5rem_1fr] gap-y-2 text-[14px]">
        {rows.map(([term, value]) => (
          <div key={term} className="contents">
            <dt className="text-muted">{term}</dt>
            <dd className="font-mono text-ink">{value}</dd>
          </div>
        ))}
      </dl>
    </section>
  );
}
