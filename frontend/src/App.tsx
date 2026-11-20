import { Link, Route, Routes } from "react-router-dom";
import { HomePage } from "./pages/HomePage";
import { ProvidersPage } from "./pages/ProvidersPage";
import { ProviderPage } from "./pages/ProviderPage";
import { ResourcePage } from "./pages/ResourcePage";
import { ModulesPage } from "./pages/ModulesPage";
import { ModulePage } from "./pages/ModulePage";
import { SiteNav } from "./components/SiteNav";
import { ThemeToggle } from "./components/ThemeToggle";
import { Mark } from "./components/Mark";
import { NotFoundPage } from "./components/States";
import { site } from "./content/site";
import { routes } from "./lib/queries";

// The shell. There is no search field in the header yet: a field that opens
// nothing is the same defect as a nav entry that links nowhere, so it arrives
// with the overlay.
export function App() {
  return (
    <div className="flex min-h-screen flex-col bg-ground text-ink">
      <header className="border-b border-line bg-surface">
        <div className="mx-auto flex max-w-shell flex-wrap items-center justify-between gap-x-6 gap-y-2 px-6 py-3">
          <Link to="/" className="flex items-center gap-2.5 text-[19px] tracking-tight">
            <Mark size={28} />
            <span className="font-semibold text-ink">{site.nameLead}</span>
            <span className="font-light text-muted">{site.nameTail}</span>
          </Link>
          <div className="flex flex-wrap items-center gap-x-4 gap-y-2">
            <SiteNav />
            {[site.links.home, site.links.docs].map((l) => (
              <a
                key={l.href}
                href={l.href}
                className="inline-flex h-11 items-center text-[15px] text-muted hover:text-ink"
              >
                {l.label}
              </a>
            ))}
            <a
              href={site.links.star.href}
              className="inline-flex h-11 items-center gap-2 rounded-box border-[1.5px] border-accent px-4 text-[15px] text-ink hover:brightness-95 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent"
            >
              <svg viewBox="0 0 24 24" width="16" height="16" aria-hidden="true" focusable="false" className="text-accent">
                <path d="M12 2.6l2.9 6 6.6.9-4.8 4.6 1.2 6.5L12 17.5l-5.9 3.1 1.2-6.5L2.5 9.5l6.6-.9z" fill="currentColor" />
              </svg>
              {site.links.star.label}
            </a>
            <ThemeToggle />
          </div>
        </div>
      </header>

      <main className="mx-auto w-full max-w-shell flex-1 px-6 py-8">
        <Routes>
          <Route path={routes.home} element={<HomePage />} />
          <Route path={routes.providers} element={<ProvidersPage />} />
          <Route path={routes.provider} element={<ProviderPage />} />
          <Route path={routes.resource} element={<ResourcePage />} />
          <Route path={routes.modules} element={<ModulesPage />} />
          <Route path={routes.module} element={<ModulePage />} />
          <Route path="*" element={<NotFoundPage />} />
        </Routes>
      </main>

      <footer className="border-t border-line bg-surface">
        <div className="mx-auto flex max-w-shell flex-wrap items-center justify-between gap-2 px-6 py-5 text-[14px] text-muted">
          <span>
            {site.name} · {site.footer}
          </span>
          <span className="flex flex-wrap gap-4">
            {site.footerLinks.map((l) => (
              <a key={l.href} href={l.href} className="hover:text-ink">
                {l.label}
              </a>
            ))}
          </span>
        </div>
      </footer>
    </div>
  );
}
