import { NavLink } from "react-router-dom";

// Sections drive both the nav and the rule that we never link a section with no
// page: adding one later is one entry here, added together with its route.
export const sections = [
  { label: "Providers", to: "/providers" },
  { label: "Modules", to: "/modules" },
];

export function SiteNav() {
  return (
    <nav aria-label="Sections" className="flex items-center gap-1">
      {sections.map((s) => (
        <NavLink
          key={s.to}
          to={s.to}
          className={({ isActive }) =>
            [
              "inline-flex h-11 items-center border-b-2 px-3 text-[15px]",
              isActive
                ? "border-warm font-semibold text-ink"
                : "border-transparent text-muted hover:text-ink",
            ].join(" ")
          }
        >
          {s.label}
        </NavLink>
      ))}
    </nav>
  );
}
