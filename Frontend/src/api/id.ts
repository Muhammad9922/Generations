/** Works in older browsers and test runtimes where crypto.randomUUID is absent. */
export function createLocalId(prefix = "local"): string {
  return `${prefix}-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 11)}`;
}
