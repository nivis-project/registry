import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";
import { Avatar } from "./Avatar";

describe("the owner avatar", () => {
  it("renders nothing at all when there is no reference", () => {
    const { container } = render(<Avatar owner="acme" />);
    // Absent, not a broken image: the contract omits the field when no avatar
    // could be obtained.
    expect(container.querySelector("img")).toBeNull();
    expect(container.textContent).toBe("");
  });

  it("names the owner for a screen reader", () => {
    render(<Avatar src="avatars/ovh.png" owner="ovh" />);
    expect(screen.getByRole("img", { name: /ovh/i })).toBeTruthy();
  });

  it("sits on a plate, so a dark transparent logo stays visible", () => {
    const { container } = render(<Avatar src="avatars/ovh.png" owner="ovh" />);
    const plate = container.firstElementChild!;
    // The plate token is light in BOTH themes; that is the whole point.
    expect(plate.className).toContain("bg-plate");
    expect(container.querySelector("img")!.parentElement).toBe(plate);
  });

  it("gives every owner the same footprint", () => {
    const { container: a } = render(<Avatar src="x.png" owner="hashicorp" />);
    const { container: b } = render(<Avatar src="y.png" owner="wearetechnative" />);
    const box = (c: HTMLElement) => (c.firstElementChild as HTMLElement).style.cssText;
    expect(box(a)).toBe(box(b));
  });

  it("is larger on a detail page than in a list, but still set by the caller", () => {
    const { container } = render(<Avatar src="x.png" owner="ovh" size={36} />);
    const plate = container.firstElementChild as HTMLElement;
    expect(plate.style.width).toBe("36px");
  });
});
