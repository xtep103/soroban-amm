import { describe, it, expect } from "vitest";
import {
  computeEfficiency,
  formatEfficiency,
  isInRange,
  rangePct,
} from "./capitalEfficiency.js";

describe("computeEfficiency", () => {
  it("returns ~1x for a full-range position", () => {
    // A very wide range (1 .. 1e9) is effectively a full-range position.
    const eff = computeEfficiency(10, 1, 1_000_000_000);
    expect(eff).toBeGreaterThan(0.999);
    expect(eff).toBeLessThan(1.01);
  });

  it("returns the hand-computed multiplier for a concentrated range", () => {
    // lower=80, upper=100: sqrt(80/100)=sqrt(0.8)=0.894427.
    // eff = 1/(1-0.894427) = 9.472135...
    const eff = computeEfficiency(90, 80, 100);
    expect(eff).toBeCloseTo(Math.sqrt(0.8) ? 1 / (1 - Math.sqrt(0.8)) : NaN, 5);
  });

  it("computes a tighter range as more efficient", () => {
    const wide = computeEfficiency(100, 50, 200);
    const tight = computeEfficiency(100, 90, 110);
    expect(tight).toBeGreaterThan(wide);
    expect(wide).toBeGreaterThan(1);
  });

  it("returns 1 for a position entirely below the current price", () => {
    expect(computeEfficiency(100, 10, 50)).toBe(1);
  });

  it("returns 1 for a position entirely above the current price", () => {
    expect(computeEfficiency(100, 150, 300)).toBe(1);
  });

  it("returns >1 when the current price straddles the range", () => {
    expect(computeEfficiency(100, 80, 120)).toBeGreaterThan(1);
  });

  it("is symmetric at the range boundaries (inclusive)", () => {
    // At exactly lower / upper the price is still "in range".
    expect(computeEfficiency(80, 80, 120)).toBeGreaterThan(1);
    expect(computeEfficiency(120, 80, 120)).toBeGreaterThan(1);
  });
});

describe("computeEfficiency — degenerate and division-by-zero inputs", () => {
  it("returns 1 for a zero lower bound", () => {
    expect(computeEfficiency(50, 0, 100)).toBe(1);
  });

  it("returns 1 for a zero upper bound", () => {
    expect(computeEfficiency(50, 10, 0)).toBe(1);
  });

  it("returns 1 when lower === upper", () => {
    expect(computeEfficiency(50, 50, 50)).toBe(1);
  });

  it("returns 1 for an inverted range (lower > upper)", () => {
    expect(computeEfficiency(50, 200, 100)).toBe(1);
  });

  it("returns 1 for a zero current price", () => {
    expect(computeEfficiency(0, 10, 100)).toBe(1);
  });

  it("returns 1 for a negative current price", () => {
    expect(computeEfficiency(-50, 10, 100)).toBe(1);
  });

  it("returns 1 for negative bounds", () => {
    expect(computeEfficiency(50, -10, 100)).toBe(1);
  });

  it("does not produce NaN/Infinity for NaN bounds", () => {
    expect(computeEfficiency(50, NaN, 100)).toBe(1);
    expect(computeEfficiency(50, 10, NaN)).toBe(1);
  });
});

describe("computeEfficiency — very large values (no overflow / precision)", () => {
  it("stays finite for huge bounds", () => {
    const eff = computeEfficiency(1e12, 1e11, 1e13);
    expect(Number.isFinite(eff)).toBe(true);
    expect(eff).toBeGreaterThanOrEqual(1);
  });

  it("returns a finite, >1 result for large concentrated bounds", () => {
    const eff = computeEfficiency(1_000_000, 900_000, 1_100_000);
    expect(Number.isFinite(eff)).toBe(true);
    expect(eff).toBeGreaterThan(1);
    expect(eff).toBeLessThan(100);
  });

  it("handles the largest safe-integer bounds without losing to Infinity", () => {
    const lower = Number.MAX_SAFE_INTEGER / 2;
    const upper = Number.MAX_SAFE_INTEGER;
    const eff = computeEfficiency((lower + upper) / 2, lower, upper);
    expect(Number.isFinite(eff)).toBe(true);
  });
});

describe("rangePct", () => {
  it("computes the percent width of a range", () => {
    expect(rangePct(80, 120, 100)).toBeCloseTo(40, 5);
  });

  it("returns 0 for a non-positive reference price (no NaN/Infinity)", () => {
    expect(rangePct(80, 120, 0)).toBe(0);
    expect(rangePct(80, 120, -10)).toBe(0);
  });

  it("handles large ranges without overflowing", () => {
    const pct = rangePct(1, Number.MAX_SAFE_INTEGER, 1);
    expect(Number.isFinite(pct)).toBe(true);
  });
});

describe("isInRange", () => {
  it("is true for a price inside the range", () => {
    expect(isInRange(100, { lower: 80, upper: 120 })).toBe(true);
  });

  it("is true at the boundaries (inclusive)", () => {
    expect(isInRange(80, { lower: 80, upper: 120 })).toBe(true);
    expect(isInRange(120, { lower: 80, upper: 120 })).toBe(true);
  });

  it("is false below and above the range", () => {
    expect(isInRange(79, { lower: 80, upper: 120 })).toBe(false);
    expect(isInRange(121, { lower: 80, upper: 120 })).toBe(false);
  });
});

describe("formatEfficiency", () => {
  it("formats a normal efficiency to one decimal", () => {
    expect(formatEfficiency(9.47)).toBe("9.5");
  });

  it("renders very large efficiencies as '>1000'", () => {
    expect(formatEfficiency(1500)).toBe(">1000");
    expect(formatEfficiency(1e9)).toBe(">1000");
  });

  it("handles the degenerate 1x case", () => {
    expect(formatEfficiency(1)).toBe("1.0");
  });
});
