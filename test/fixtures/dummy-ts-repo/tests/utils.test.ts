import { describe, it, expect } from "vitest";
import { formatCurrency, truncate, slugify } from "../src/utils/format";
import { isValidEmail, clamp, isPositive } from "../src/utils/validate";

describe("format", () => {
  it("formats currency", () => {
    expect(formatCurrency(1234.5)).toContain("1,234.50");
  });
  it("truncates", () => {
    expect(truncate("hello world", 8)).toBe("hello w…");
  });
  it("slugifies", () => {
    expect(slugify("Hello, World!")).toBe("hello-world");
  });
});

describe("validate", () => {
  it("recognises valid emails", () => {
    expect(isValidEmail("alice@example.test")).toBe(true);
    expect(isValidEmail("not-an-email")).toBe(false);
  });
  it("clamps", () => {
    expect(clamp(5, 0, 3)).toBe(3);
    expect(clamp(-1, 0, 3)).toBe(0);
    expect(clamp(2, 0, 3)).toBe(2);
  });
  it("isPositive", () => {
    expect(isPositive(1)).toBe(true);
    expect(isPositive(0)).toBe(false);
    expect(isPositive(NaN)).toBe(false);
  });
});
