import { Link, Route, Routes } from "react-router-dom";
import { HomePage } from "./pages/HomePage";
import { ProvidersPage } from "./pages/ProvidersPage";
import { ProviderPage } from "./pages/ProviderPage";
import { ResourcePage } from "./pages/ResourcePage";
import { SiteNav } from "./components/SiteNav";

// App is the data-agnostic SPA shell. The root route orients a visitor; each
// section owns its own route below it. Resource-page rendering is the adapted
// part of the registry-ui model.
export function App() {
  return (
    <div className="min-h-screen bg-slate-50">
      <header className="border-b border-slate-200 bg-white">
        <div className="mx-auto flex max-w-5xl items-center justify-between px-4 py-3">
          <Link to="/" className="text-lg font-bold text-slate-900">
            ❄ Nivis Registry
          </Link>
          <SiteNav />
        </div>
      </header>
      <main className="mx-auto max-w-5xl px-4 py-6">
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/providers" element={<ProvidersPage />} />
          <Route
            path="/providers/:namespace/:name/:version"
            element={<ProviderPage />}
          />
          <Route
            path="/providers/:namespace/:name/:version/resources/:item"
            element={<ResourcePage />}
          />
        </Routes>
      </main>
    </div>
  );
}
