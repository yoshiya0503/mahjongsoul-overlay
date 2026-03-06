(() => {
  "use strict";

  const WS_URL = `${location.protocol === 'https:' ? 'wss:' : 'ws:'}//${location.host}/ws/overlay`;
  const WIND = ["東", "南", "西", "北"];

  let ws = null;
  let prevState = null;
  let reconnectTimer = null;
  let pingTimer = null;

  // --- WebSocket ---
  function connect() {
    ws = new WebSocket(WS_URL);
    ws.onopen = () => {
      clearTimeout(reconnectTimer);
      clearInterval(pingTimer);
      pingTimer = setInterval(() => {
        if (ws && ws.readyState === 1) ws.send("ping");
      }, 30000);
    };
    ws.onmessage = (event) => {
      try {
        const state = JSON.parse(event.data);
        render(state);
      } catch (e) {
      }
    };
    ws.onclose = () => {
      clearInterval(pingTimer);
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
    if (!state.inGame) {
      for (let i = 0; i < 4; i++) {
        const row = document.getElementById(`player-${i}`);
        if (!row) continue;
        row.className = "player-row";
        if (!row.dataset.initialized) {
          row.innerHTML = `
            <div class="player-rank"></div>
            <div class="player-name"></div>
            <div class="player-score"></div>
          `;
          row.dataset.initialized = "1";
        }
        row.querySelector(".player-rank").textContent = "-";
        row.querySelector(".player-name").textContent = "-";
        row.querySelector(".player-score").textContent = "-";
      }
      return;
    }

    const sorted = [...state.players].sort((a, b) => a.rank - b.rank);

    sorted.forEach((player, i) => {
      const row = document.getElementById(`player-${i}`);
      if (!row) return;

      row.className = "player-row rank-" + player.rank;

      // 初回のみDOM構築、以降は差分更新
      if (!row.dataset.initialized) {
        row.innerHTML = `
          <div class="player-rank"></div>
          <div class="player-name"></div>
          <div class="player-score"></div>
        `;
        row.dataset.initialized = "1";
      }

      row.querySelector(".player-rank").textContent = player.rank;
      row.querySelector(".player-name").textContent = player.name;
      row.querySelector(".player-score").textContent = formatScore(player.score);

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

    // プレースホルダー行
    for (let i = sorted.length; i < 4; i++) {
      const row = document.getElementById(`player-${i}`);
      if (!row) continue;
      row.className = "player-row";
      if (!row.dataset.initialized) {
        row.innerHTML = `
          <div class="player-rank"></div>
          <div class="player-name"></div>
          <div class="player-score"></div>
        `;
        row.dataset.initialized = "1";
      }
      row.querySelector(".player-rank").textContent = "-";
      row.querySelector(".player-name").textContent = "-";
      row.querySelector(".player-score").textContent = "-";
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
    const history = state.history || [];
    const container = document.getElementById("score-toast-container");
    const prevCount = prevState?.history?.length || 0;

    // 新しい結果がない場合はスキップ
    if (history.length <= prevCount) return;

    // 全履歴を再描画（初回接続時も含む）
    if (prevCount === 0 && history.length > 0) {
      container.innerHTML = "";
      for (const result of history) {
        appendResultEntry(container, result, state);
      }
    } else {
      // 差分だけ追加
      const newResults = history.slice(prevCount);
      for (const result of newResults) {
        appendResultEntry(container, result, state);
      }
    }
  }

  function appendResultEntry(container, result, state) {
    const r = result.round || {};
    const WIND = ["東", "南", "西", "北"];
    const roundLabel = `${WIND[r.chang] || "?"}${(r.ju || 0) + 1}局`;

    let detail = "";
    const changes = (result.scoreChanges || []).filter((sc) => sc.delta !== 0);
    if (changes.length > 0) {
      detail = changes.map((sc) => {
        const player = state.players.find((p) => p.seat === sc.seat);
        const name = player ? player.name : `P${sc.seat}`;
        const sign = sc.delta > 0 ? "+" : "";
        return `${escapeHtml(name)}: ${sign}${sc.delta}`;
      }).join("  ");
    } else {
      detail = result.winner >= 0 ? "和了" : "流局";
    }

    const entry = document.createElement("div");
    entry.className = "score-entry";
    entry.innerHTML = `<span class="score-entry-round">${roundLabel}</span> ${detail}`;
    container.appendChild(entry);
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

  // --- Session Clear ---
  document.getElementById("session-clear").addEventListener("click", () => {
    fetch("/api/session/clear", { method: "POST" });
  });

  // --- Init ---
  connect();
})();
