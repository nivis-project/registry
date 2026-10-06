import { NavLink } from "react-router-dom";

// Sections drive both the nav and the rule that we never link a section that
// has no page: adding Modules later is one entry here, added together with its
// route.
export const sections = [
  { label: "Providers", to: "/providers" },
  { label: "Modules", to: "/modules" },
];

export function SiteNav() {
  return (
    <nav className="flex items-center gap-1">
      {sections.map((s) => (
        <NavLink
          key={s.to}
          to={s.to}
          className={({ isActive }) =>
            [
              "rounded-md px-3 py-1.5 text-sm font-medium",
              isActive
                ? "bg-sky-100 text-sky-900"
                : "text-slate-600 hover:bg-slate-100 hover:text-slate-900",
            ].join(" ")
          }
        >
          {s.label}
        </NavLink>
      ))}
    </nav>
  );
}
