export function formatNumber(value: number, digits = 4): string {
  if (!Number.isFinite(value)) {
    return value > 0 ? "∞" : "−∞";
  }
  if (value !== 0 && (Math.abs(value) >= 1e6 || Math.abs(value) < 1e-4)) {
    return value.toExponential(3).replace("e", "·10^");
  }
  return value
    .toFixed(digits)
    .replace(/(\.\d*?[1-9])0+$/, "$1")
    .replace(/\.0+$/, "");
}

export function formatPValue(value: number): string {
  if (!Number.isFinite(value)) {
    return value > 0 ? "∞" : "−∞";
  }
  if (value < 1e-4) {
    return value.toExponential(2);
  }
  return value.toFixed(4);
}

export function formatCount(value: number): string {
  return value.toLocaleString("ru-RU", { maximumFractionDigits: 1 });
}

export function formatPercent(value: number, digits = 2): string {
  return `${(value * 100).toFixed(digits)}%`;
}
