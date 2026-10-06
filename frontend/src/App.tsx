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
            <a
              href={site.homeUrl}
              className="inline-flex h-11 items-center text-[15px] text-muted hover:text-ink"
            >
              nivis.tf
            </a>
            <ThemeToggle />
          </div>
        </div>
      </header>

      <main className="mx-auto w-full max-w-shell flex-1 px-6 py-8">
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/providers" element={<ProvidersPage />} />
          <Route path="/providers/:namespace/:name/:version" element={<ProviderPage />} />
          <Route
            path="/providers/:namespace/:name/:version/resources/:item"
            element={<ResourcePage />}
          />
          <Route path="/modules" element={<ModulesPage />} />
          <Route path="/modules/:owner/:name/:version" element={<ModulePage />} />
          <Route path="*" element={<NotFoundPage />} />
        </Routes>
      </main>

      <footer className="border-t border-line bg-surface">
        <div className="mx-auto flex max-w-shell flex-wrap items-center justify-between gap-2 px-6 py-5 text-[14px] text-muted">
          <span>
            {site.name} · {site.footer}
          </span>
          <span className="flex gap-4">
            <a href={site.homeUrl} className="hover:text-ink">
              nivis.tf
            </a>
            <a href={site.repoUrl} className="hover:text-ink">
              GitHub
            </a>
          </span>
        </div>
      </footer>
    </div>
  );
}
