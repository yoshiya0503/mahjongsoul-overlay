export const WIND = ["東", "南", "西", "北"];

export function formatScore(score) {
  if (score === undefined || score === null) return "-";
  return score.toLocaleString();
}

export function escapeHtml(str) {
  const div = document.createElement("div");
  div.textContent = str;
  return div.innerHTML;
}
