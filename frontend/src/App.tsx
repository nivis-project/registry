import { Route, Routes } from "react-router-dom";
import { IndexPage } from "./pages/IndexPage";
import { ProviderPage } from "./pages/ProviderPage";
import { ResourcePage } from "./pages/ResourcePage";

// App is the data-agnostic SPA shell: routing kept from the registry-ui model
// (index → provider → item). The resource-page rendering is the adapted part.
export function App() {
  return (
    <div className="min-h-screen bg-slate-50">
      <header className="border-b border-slate-200 bg-white">
        <div className="mx-auto max-w-5xl px-4 py-3">
          <a href="." className="text-lg font-bold text-slate-900">
            ❄ Nivis Registry
          </a>
        </div>
      </header>
      <main className="mx-auto max-w-5xl px-4 py-6">
        <Routes>
          <Route path="/" element={<IndexPage />} />
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
