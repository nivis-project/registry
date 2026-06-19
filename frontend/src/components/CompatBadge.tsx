import type { CompatRecord } from "../lib/contract";

// CompatBadge surfaces the four-axis compatibility record honestly: the headline
// tier ("compatible by design" vs "e2e verified"), the protocol(s), the
// Nivis-supported ∩ published architectures, and the e2e status. It never shows
// "verified" unless the contract says so.
export function CompatBadge({ compat }: { compat: CompatRecord }) {
  const verified = compat.tier === "e2e verified" || compat.e2e === "verified";
  return (
    <div className="rounded-lg border border-slate-200 bg-white p-4">
      <div className="flex items-center gap-2">
        <span
          className={
            "inline-flex items-center rounded-full px-3 py-1 text-sm font-semibold " +
            (verified
              ? "bg-emerald-100 text-emerald-800"
              : "bg-sky-100 text-sky-800")
          }
        >
          {verified ? "✓ e2e verified" : "compatible by design"}
        </span>
        {compat.schema_extractable && (
          <span className="text-xs text-slate-500">schema-extractable</span>
        )}
      </div>
      <dl className="mt-3 grid grid-cols-[8rem_1fr] gap-y-1 text-sm">
        <dt className="text-slate-500">Protocol</dt>
        <dd className="text-slate-800">{compat.protocols.join(", ") || "—"}</dd>
        <dt className="text-slate-500">Architectures</dt>
        <dd className="text-slate-800">{compat.architectures.join(", ") || "—"}</dd>
        <dt className="text-slate-500">E2E</dt>
        <dd className="text-slate-800">{compat.e2e}</dd>
      </dl>
    </div>
  );
}
