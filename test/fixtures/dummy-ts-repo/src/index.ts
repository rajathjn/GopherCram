import { fetchUsers } from "./api/users";
import { fetchProducts } from "./api/products";
import { formatCurrency } from "./utils/format";

async function main(): Promise<void> {
  const users = await fetchUsers();
  const products = await fetchProducts();
  console.log(`loaded ${users.length} users, ${products.length} products`);
  for (const p of products) {
    console.log(`- ${p.name}: ${formatCurrency(p.price)}`);
  }
}

main().catch((err) => {
  console.error("fatal:", err);
  process.exit(1);
});
