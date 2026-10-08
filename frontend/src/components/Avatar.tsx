// The owner's brand avatar.
//
// Always on a neutral plate. Seven of the thirty-six real avatars are dark ink
// on transparency and would simply disappear on a dark background, and two of
// those belong to this project. The plate also makes marks of wildly different
// shape, padding and background read as one row.
//
// An avatar says who publishes a thing. It is not a quality mark, so every
// owner gets the same footprint and the same treatment.
//
// `src` is the CONTRACT-RELATIVE reference ("avatars/ovh.png"). Resolving it
// against the contract root is this component's job: handing it to <img>
// unchanged resolves it against the document and 404s.
import { avatarURL } from "../lib/data";

export function Avatar({
  src,
  owner,
  size = 28,
}: {
  src?: string;
  owner: string;
  size?: number;
}) {
  const url = avatarURL(src);
  if (!url) return null;
  return (
    <span
      className="inline-flex shrink-0 items-center justify-center overflow-hidden rounded-[6px] border border-line bg-plate"
      style={{ width: size, height: size }}
    >
      <img
        src={url}
        alt={`${owner} logo`}
        width={size - 6}
        height={size - 6}
        loading="lazy"
        decoding="async"
        className="h-[calc(100%-6px)] w-[calc(100%-6px)] object-contain"
      />
    </span>
  );
}
