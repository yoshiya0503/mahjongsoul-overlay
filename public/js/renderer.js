import { WIND, formatScore, escapeHtml } from "./utils.js";

let prevState = null;

export function render(state) {
  const overlay = document.getElementById("overlay");
  overlay.classList.remove("hidden");
  overlay.classList.add("visible");

  renderRound(state);
  renderScoreboard(state);
  renderRankPoint(state);
  renderSession(state);
  renderScoreToasts(state);

  prevState = structuredClone(state);
}

function renderRound(state) {
  if (!state.inGame) {
    document.getElementById("round-text").textContent = "WAITING";
    document.getElementById("honba-text").textContent = "";
    return;
  }
  const r = state.round;
  document.getElementById("round-text").textContent =
    `${WIND[r.chang] || "?"}${r.ju + 1}局`;
  document.getElementById("honba-text").textContent =
    r.ben > 0 ? `${r.ben}本場` : "";
}

function initPlayerRow(row) {
  if (row.dataset.initialized) return;
  row.innerHTML = `
    <div class="player-rank"></div>
    <div class="player-name"></div>
    <div class="player-score"></div>
  `;
  row.dataset.initialized = "1";
}

function setPlayerRow(row, rank, name, score) {
  row.querySelector(".player-rank").textContent = rank;
  row.querySelector(".player-name").textContent = name;
  row.querySelector(".player-score").textContent = score;
}

function renderScoreboard(state) {
  if (!state.inGame) {
    for (let i = 0; i < 4; i++) {
      const row = document.getElementById(`player-${i}`);
      if (!row) continue;
      row.className = "player-row";
      initPlayerRow(row);
      setPlayerRow(row, "-", "-", "-");
    }
    return;
  }

  const sorted = [...state.players].sort((a, b) => a.rank - b.rank);

  sorted.forEach((player, i) => {
    const row = document.getElementById(`player-${i}`);
    if (!row) return;

    row.className = "player-row rank-" + player.rank;
    initPlayerRow(row);
    setPlayerRow(row, player.rank, player.name, formatScore(player.score));

    if (prevState && prevState.players) {
      const prev = prevState.players.find((p) => p.seat === player.seat);
      if (prev && prev.score !== player.score) {
        const delta = player.score - prev.score;
        const deltaEl = document.createElement("div");
        deltaEl.className = `score-delta ${delta > 0 ? "positive" : "negative"}`;
        deltaEl.textContent = delta > 0 ? `+${delta}` : `${delta}`;
        row.appendChild(deltaEl);

        const scoreEl = row.querySelector(".player-score");
        scoreEl.classList.add(delta > 0 ? "score-up" : "score-down");
        setTimeout(() => {
          scoreEl.classList.remove("score-up", "score-down");
        }, 2000);
      }
    }
  });

  for (let i = sorted.length; i < 4; i++) {
    const row = document.getElementById(`player-${i}`);
    if (!row) continue;
    row.className = "player-row";
    initPlayerRow(row);
    setPlayerRow(row, "-", "-", "-");
  }
}

function renderRankPoint(state) {
  const rp = state.rankPoint;
  document.getElementById("rank-name").textContent = rp?.rankName || "-";
  const pct = rp?.targetPt > 0 ? (rp.currentPt / rp.targetPt) * 100 : 0;
  document.getElementById("rank-bar").style.width = `${Math.min(pct, 100)}%`;
  document.getElementById("rank-pt-text").textContent =
    rp?.targetPt ? `${rp.currentPt} / ${rp.targetPt} pt` : "- / - pt";
}

function renderSession(state) {
  const session = state.session || [];

  document.getElementById("session-games").textContent = session.length;

  if (session.length > 0) {
    const avgRank = session.reduce((s, r) => s + r.rank, 0) / session.length;
    document.getElementById("session-avg").textContent = avgRank.toFixed(2);

    const firstCount = session.filter((r) => r.rank === 1).length;
    const firstRate = ((firstCount / session.length) * 100).toFixed(0);
    document.getElementById("session-first").textContent = `${firstRate}%`;

    const totalPt = session.reduce((s, r) => s + r.deltaPt, 0);
    const ptEl = document.getElementById("session-pt");
    ptEl.textContent = totalPt >= 0 ? `+${totalPt}` : `${totalPt}`;
    ptEl.style.color = totalPt >= 0 ? "#69f0ae" : "#ef5350";
  } else {
    document.getElementById("session-avg").textContent = "-";
    document.getElementById("session-first").textContent = "-";
    const ptEl = document.getElementById("session-pt");
    ptEl.textContent = "±0";
    ptEl.style.color = "";
  }

  const historyEl = document.getElementById("session-history");
  const prevCount = prevState?.session?.length || 0;
  if (session.length !== prevCount) {
    historyEl.innerHTML = session
      .map((r) => `<div class="session-badge rank-${r.rank}">${r.rank}</div>`)
      .join("");
  }
}

function renderScoreToasts(state) {
  const history = state.history || [];
  const container = document.getElementById("score-toast-container");
  const prevCount = prevState?.history?.length || 0;

  if (history.length <= prevCount) return;

  if (prevCount === 0 && history.length > 0) {
    container.innerHTML = "";
    for (const result of history) {
      appendResultEntry(container, result, state);
    }
  } else {
    const newResults = history.slice(prevCount);
    for (const result of newResults) {
      appendResultEntry(container, result, state);
    }
  }
}

function appendResultEntry(container, result, state) {
  const r = result.round || {};
  const roundLabel = `${WIND[r.chang] || "?"}${(r.ju || 0) + 1}局`;

  let detail = "";
  const changes = (result.scoreChanges || []).filter((sc) => sc.delta !== 0);
  if (changes.length > 0) {
    detail = changes
      .map((sc) => {
        const player = state.players.find((p) => p.seat === sc.seat);
        const name = player ? player.name : `P${sc.seat}`;
        const sign = sc.delta > 0 ? "+" : "";
        return `${escapeHtml(name)}: ${sign}${sc.delta}`;
      })
      .join("  ");
  } else {
    detail = result.winner >= 0 ? "和了" : "流局";
  }

  const entry = document.createElement("div");
  entry.className = "score-entry";
  entry.innerHTML = `<span class="score-entry-round">${roundLabel}</span> ${detail}`;
  container.appendChild(entry);
}
