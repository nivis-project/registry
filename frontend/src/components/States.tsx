import { Link } from "react-router-dom";
import { states } from "../content/site";
import { useDocumentTitle } from "../lib/useDocumentTitle";

// Skeletons occupy the eventual layout, so arriving content shifts nothing.
export function SkeletonRows({ rows = 6 }: { rows?: number }) {
  return (
    <ul
      className="divide-y divide-line overflow-hidden rounded-card border border-line bg-surface"
      aria-busy="true"
      aria-label={states.loading}
    >
      {Array.from({ length: rows }).map((_, i) => (
        <li key={i} className="flex items-center justify-between gap-4 px-4 py-4">
          <span className="h-4 w-1/3 rounded bg-line" />
          <span className="h-4 w-16 rounded bg-line" />
        </li>
      ))}
    </ul>
  );
}

export function ErrorState({ what, error, onRetry }: { what: string; error: unknown; onRetry?: () => void }) {
  return (
    <div role="alert" className="rounded-card border border-danger-line bg-danger-bg p-4 text-danger-ink">
      <p className="font-medium">{states.errorTitle}</p>
      <p className="mt-1 text-[15px]">{what}</p>
      <p className="mt-1 font-mono text-[13px] opacity-90">{String(error)}</p>
      {onRetry && (
        <button
          type="button"
          onClick={onRetry}
          className="mt-3 h-11 rounded-box border border-danger-line px-4 font-medium"
        >
          {states.retry}
        </button>
      )}
    </div>
  );
}

export function EmptyState({ children }: { children: React.ReactNode }) {
  return <p className="rounded-card border border-line bg-surface px-4 py-6 text-muted">{children}</p>;
}

export function NotFoundPage() {
  useDocumentTitle(states.notFoundTitle);
  return (
    <div>
      <h1 className="text-[clamp(28px,3.4vw,40px)] font-semibold text-ink">{states.notFoundTitle}</h1>
      <p className="mt-2 text-muted">{states.notFoundBody}</p>
      <Link to="/" className="mt-4 inline-block font-medium text-accent hover:underline">
        {states.backHome}
      </Link>
    </div>
  );
}
