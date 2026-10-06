// The Nivis mark, drawn from its defining formula rather than stored as an
// asset, so the header and the favicon cannot drift apart.
//
//   r(θ) = a + b·cos(kθ),  b = a·amp/(k²+1)
//
// The registry variant is k = 4, amp = 1.2, in three layers.

const K = 4;
const AMP = 1.2;
const SAMPLES = 120;

function petalPath(a: number, rotationDeg: number): string {
  const b = (a * AMP) / (K * K + 1);
  const phase = (rotationDeg * Math.PI) / 180;
  let d = "";
  for (let i = 0; i < SAMPLES; i++) {
    const t = (i / SAMPLES) * Math.PI * 2;
    const r = a + b * Math.cos(K * t);
    const x = r * Math.cos(t + phase);
    const y = r * Math.sin(t + phase);
    d += `${i === 0 ? "M" : "L"}${x.toFixed(2)} ${y.toFixed(2)}`;
  }
  return `${d}Z`;
}

// Computed once for the module, not once per render.
const LAYERS = [
  { d: petalPath(84, 29), className: "fill-mark-a", opacity: 0.35 },
  { d: petalPath(80, 61), className: "fill-mark-b", opacity: 0.35 },
  { d: petalPath(64, 45), className: "fill-mark-core", opacity: 1 },
] as const;

export function Mark({
  size = 28,
  title,
}: {
  size?: number;
  title?: string;
}) {
  return (
    <svg
      viewBox="-100 -100 200 200"
      width={size}
      height={size}
      role={title ? "img" : undefined}
      aria-hidden={title ? undefined : true}
      aria-label={title}
    >
      {title && <title>{title}</title>}
      {LAYERS.map((l, i) => (
        <path key={i} d={l.d} className={l.className} opacity={l.opacity} />
      ))}
    </svg>
  );
}

// markPaths is exported for the favicon, which needs the same geometry outside
// of React.
export const markPaths = LAYERS.map((l) => ({ d: l.d, opacity: l.opacity }));
