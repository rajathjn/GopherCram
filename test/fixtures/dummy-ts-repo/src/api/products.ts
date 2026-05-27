export interface Product {
  sku: string;
  name: string;
  price: number;
}

export async function fetchProducts(): Promise<Product[]> {
  await new Promise((r) => setTimeout(r, 1));
  return [
    { sku: "A100", name: "Widget", price: 9.99 },
    { sku: "A200", name: "Gadget", price: 19.99 },
    { sku: "A300", name: "Gizmo", price: 29.99 },
  ];
}

export function totalPrice(products: Product[]): number {
  return products.reduce((sum, p) => sum + p.price, 0);
}
