const EMAIL_RE = /^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$/i;

export function isValidEmail(input: string): boolean {
  return EMAIL_RE.test(input.trim());
}

export function clamp(n: number, min: number, max: number): number {
  if (n < min) return min;
  if (n > max) return max;
  return n;
}

export function isPositive(n: number): boolean {
  return Number.isFinite(n) && n > 0;
}
