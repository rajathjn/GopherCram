import { describe, it, expect } from "vitest";
import { fetchUsers, findUserByEmail } from "../src/api/users";
import { fetchProducts, totalPrice } from "../src/api/products";

describe("users api", () => {
  it("returns users", async () => {
    const users = await fetchUsers();
    expect(users.length).toBeGreaterThan(0);
  });

  it("finds by email case-insensitively", async () => {
    const users = await fetchUsers();
    const u = findUserByEmail(users, "ADA@example.test");
    expect(u?.name).toBe("Ada Lovelace");
  });
});

describe("products api", () => {
  it("sums prices", async () => {
    const products = await fetchProducts();
    expect(totalPrice(products)).toBeCloseTo(59.97, 2);
  });
});
