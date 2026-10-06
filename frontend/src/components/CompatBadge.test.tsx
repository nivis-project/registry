import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";
import { CompatBadge } from "./CompatBadge";
import type { CompatRecord } from "../lib/contract";

const base: CompatRecord = {
  address: "hashicorp/random",
  tier: "compatible by design",
  schema_extractable: true,
  protocols: ["5.0"],
  architectures: ["linux/amd64"],
  e2e: "none",
};

describe("the compatibility panel", () => {
  it("shows the default tier, not verified", () => {
    render(<CompatBadge compat={base} />);
    expect(screen.getByText(/^compatible by design$/i)).toBeTruthy();
    expect(screen.queryByText(/verified/i)).toBeNull();
  });

  it("shows verified only when the contract says so", () => {
    render(<CompatBadge compat={{ ...base, e2e: "verified" }} />);
    expect(screen.getByText(/e2e verified/i)).toBeTruthy();
  });

  it("falls back to a dash for every empty axis", () => {
    render(
      <CompatBadge
        compat={{ ...base, schema_extractable: false, protocols: [], architectures: [], e2e: "" as never }}
      />,
    );
    // Schema, Protocol, Systems and End to end all unknown.
    expect(screen.getAllByText("-")).toHaveLength(4);
  });

  it("reports the axes it does have", () => {
    render(<CompatBadge compat={{ ...base, architectures: ["linux/amd64", "darwin/arm64"] }} />);
    expect(screen.getByText("linux/amd64, darwin/arm64")).toBeTruthy();
    expect(screen.getByText("5.0")).toBeTruthy();
  });
});
