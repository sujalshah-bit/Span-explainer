export function hasToken(): boolean {
  return Boolean(localStorage.getItem("token"));
}