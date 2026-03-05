(() => {
  "use strict";

  const WS_URL = `${location.protocol === 'https:' ? 'wss:' : 'ws:'}//${location.host}/ws/overlay`;
  const WIND = ["東", "南", "西", "北"];

  let ws = null;
  let prevState = null;
  let reconnectTimer = null;

  // --- WebSocket ---
  function connect() {
    ws = new WebSocket(WS_URL);
    ws.onopen = () => {
      console.log("[overlay] connected");
      clearTimeout(reconnectTimer);
    };
    ws.onmessage = (event) => {
      try {
        const state = JSON.parse(event.data);
        render(state);
      } catch (e) {
        console.error("[overlay] parse error:", e);
      }
    };
    ws.onclose = () => {
      console.log("[overlay] disconnected");
      reconnectTimer = setTimeout(connect, 2000);
    };
  }

  // --- Render ---
  function render(state) {
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
    const windChar = WIND[r.chang] || "?";
    const roundNum = r.ju + 1;
    document.getElementById("round-text").textContent =
      `${windChar}${roundNum}局`;
    document.getElementById("honba-text").textContent =
      r.ben > 0 ? `${r.ben}本場` : "";
  }

  function renderScoreboard(state) {
    const sorted = [...state.players].sort((a, b) => a.rank - b.rank);

    sorted.forEach((player, i) => {
      const row = document.getElementById(`player-${i}`);
      if (!row) return;

      // Clear old rank classes
      row.className = "player-row rank-" + player.rank;

      row.innerHTML = `
        <div class="player-rank">${player.rank}</div>
        <div class="player-name">${escapeHtml(player.name)}</div>
        <div class="player-score">${formatScore(player.score)}</div>
      `;

      // Score change animation
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

    // Hide unused rows
    for (let i = sorted.length; i < 4; i++) {
      const row = document.getElementById(`player-${i}`);
      if (row) row.innerHTML = "";
    }
  }

  function renderRankPoint(state) {
    const rp = state.rankPoint;
    const container = document.getElementById("rank-point");

    if (!rp || !rp.rankName) {
      container.style.display = "none";
      return;
    }
    container.style.display = "";

    document.getElementById("rank-name").textContent = rp.rankName;
    const pct = rp.targetPt > 0 ? (rp.currentPt / rp.targetPt) * 100 : 0;
    document.getElementById("rank-bar").style.width = `${Math.min(pct, 100)}%`;
    document.getElementById("rank-pt-text").textContent =
      `${rp.currentPt} / ${rp.targetPt} pt`;
  }

  function renderSession(state) {
    const session = state.session || [];
    const container = document.getElementById("session");

    if (session.length === 0) {
      container.style.display = "none";
      return;
    }
    container.style.display = "";

    document.getElementById("session-games").textContent = session.length;

    const avgRank = session.reduce((s, r) => s + r.rank, 0) / session.length;
    document.getElementById("session-avg").textContent = avgRank.toFixed(2);

    const firstCount = session.filter((r) => r.rank === 1).length;
    const firstRate = ((firstCount / session.length) * 100).toFixed(0);
    document.getElementById("session-first").textContent = `${firstRate}%`;

    const totalPt = session.reduce((s, r) => s + r.deltaPt, 0);
    const ptEl = document.getElementById("session-pt");
    ptEl.textContent = totalPt >= 0 ? `+${totalPt}` : `${totalPt}`;
    ptEl.style.color = totalPt >= 0 ? "#69f0ae" : "#ef5350";

    // History badges
    const historyEl = document.getElementById("session-history");
    const prevCount = prevState?.session?.length || 0;
    if (session.length !== prevCount) {
      historyEl.innerHTML = session
        .map(
          (r) =>
            `<div class="session-badge rank-${r.rank}">${r.rank}</div>`
        )
        .join("");
    }
  }

  function renderScoreToasts(state) {
    if (!prevState || !state.history || !prevState.history) return;
    if (state.history.length <= prevState.history.length) return;

    const newResults = state.history.slice(prevState.history.length);
    const container = document.getElementById("score-toast-container");

    for (const result of newResults) {
      if (!result.scoreChanges || result.scoreChanges.length === 0) continue;

      const lines = result.scoreChanges
        .filter((sc) => sc.delta !== 0)
        .map((sc) => {
          const player = state.players.find((p) => p.seat === sc.seat);
          const name = player ? player.name : `P${sc.seat}`;
          const sign = sc.delta > 0 ? "+" : "";
          return `${name}: ${sign}${sc.delta}`;
        });

      if (lines.length === 0) continue;

      const toast = document.createElement("div");
      toast.className = "score-toast";
      toast.textContent = lines.join("  ");
      container.appendChild(toast);

      setTimeout(() => toast.remove(), 4000);
    }
  }

  // --- Utils ---
  function formatScore(score) {
    if (score === undefined || score === null) return "-";
    return score.toLocaleString();
  }

  function escapeHtml(str) {
    const div = document.createElement("div");
    div.textContent = str;
    return div.innerHTML;
  }

  // --- Init ---
  connect();
})();
