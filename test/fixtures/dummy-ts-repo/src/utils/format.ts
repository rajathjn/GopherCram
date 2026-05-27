export function formatCurrency(amount: number, locale = "en-US", currency = "USD"): string {
  return new Intl.NumberFormat(locale, { style: "currency", currency }).format(amount);
}

export function truncate(input: string, max: number, ellipsis = "…"): string {
  if (input.length <= max) return input;
  return input.slice(0, Math.max(0, max - ellipsis.length)) + ellipsis;
}

export function slugify(input: string): string {
  return input
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");
}
