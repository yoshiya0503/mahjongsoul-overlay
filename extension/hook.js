// 雀魂 WebSocket Hook - ゲーム内部のprotobuf.jsデコード結果をフックする
(function () {
  "use strict";

  if (window.__mjsHookInstalled) return;
  window.__mjsHookInstalled = true;

  const SERVER_URL = "wss://mahjongsoul-overlay.fly.dev/ws/hook";

  let hookSocket = null;
  const OrigWebSocket = window.WebSocket;


  // --- サーバー接続 ---
  function connectServer() {
    if (hookSocket && hookSocket.readyState <= 1) return;
    try {
      hookSocket = new OrigWebSocket(SERVER_URL);
      hookSocket.onopen = () => {};
      hookSocket.onclose = () => setTimeout(connectServer, 5000);
      hookSocket.onerror = () => {};
    } catch (e) {}
  }

  function sendToServer(type, data) {
    try {
      if (hookSocket && hookSocket.readyState === 1) {
        hookSocket.send(JSON.stringify({ type, data }));
      }
    } catch (e) {}
  }

  // --- 段位マッピング ---
  // level.id: 上2桁=モード(10=四麻,20=三麻), 次2桁=段位, 最後1桁=星
  const RANK_NAMES = {
    1: "初心", 2: "雀士", 3: "雀傑", 4: "雀豪", 5: "雀聖", 6: "魂天"
  };
  const RANK_TARGETS = {
    101: 20, 102: 80, 103: 200,     // 初心1-3
    201: 600, 202: 800, 203: 1000,  // 雀士1-3
    301: 1200, 302: 1400, 303: 2000, // 雀傑1-3
    401: 2800, 402: 3200, 403: 3600, // 雀豪1-3
    501: 4000, 502: 6000, 503: 9000, // 雀聖1-3
    601: 0, // 魂天
  };

  function getRankName(levelId) {
    // 10503 → mode=10, rank=5, star=03 → 雀聖3
    const rank = Math.floor((levelId % 10000) / 100);
    const star = levelId % 100;
    return `${RANK_NAMES[rank] || "?"}${star}`;
  }

  function getRankTarget(levelId) {
    const key = levelId % 10000; // remove mode prefix
    return RANK_TARGETS[key] || 0;
  }

  // --- protobuf.js デコードフック ---
  const WATCHED = new Set([
    "ResAuthGame", "ResEnterGame",
    "ActionNewRound", "ActionHule", "ActionNoTile", "ActionLiuJu",
    "NotifyGameEndResult", "GameEndResult"
  ]);


  let protoHooked = false;

  function hookProtobuf(pb) {
    if (protoHooked) return;
    if (!pb || !pb.Type || !pb.Type.prototype || !pb.Type.prototype.decode) return;
    protoHooked = true;

    const origDecode = pb.Type.prototype.decode;
    pb.Type.prototype.decode = function (reader, length) {
      const msg = origDecode.call(this, reader, length);
      try {
        if (WATCHED.has(this.name)) {
          const obj = this.toObject(msg, { defaults: true, longs: Number, enums: Number });
          handleDecoded(this.name, obj);
        }
      } catch (e) {}
      return msg;
    };
  }

  function handleDecoded(name, obj) {
    switch (name) {
      case "ResAuthGame":
      case "ResEnterGame": {
        const allPlayers = [...(obj.players || []), ...(obj.robots || [])];
        const seatList = obj.seat_list || [];
        const seatMap = {};
        for (const p of allPlayers) {
          seatMap[p.account_id] = p;
        }

        const ordered = seatList.map((accountId, seat) => {
          const p = seatMap[accountId] || {};
          return {
            name: p.nickname || `CPU ${seat + 1}`,
            character: String(p.character?.charid || ""),
          };
        });
        if (ordered.length > 0) sendToServer("authGame", { players: ordered });

        // 段位ポイント
        const me = (obj.players || [])[0];
        if (me && me.level) {
          sendToServer("rankPoint", {
            currentPt: me.level.score || 0,
            targetPt: getRankTarget(me.level.id),
            rankName: getRankName(me.level.id),
          });
        }
        break;
      }
      case "ActionNewRound": {
        sendToServer("newRound", {
          chang: obj.chang || 0,
          ju: obj.ju || 0,
          ben: obj.ben || 0,
          scores: obj.scores || [],
        });
        break;
      }
      case "ActionHule": {
        const scores = obj.scores || [];
        const deltaScores = obj.delta_scores || obj.deltaScores || [];
        let winner = -1;
        if (obj.hules && obj.hules.length > 0) {
          winner = obj.hules[0].seat || 0;
        }
        sendToServer("hule", { scores, deltaScores, winner });
        break;
      }
      case "ActionNoTile": {
        const scores = obj.scores || [];
        const deltaScores = obj.delta_scores || obj.deltaScores || [];
        sendToServer("noTile", { scores, deltaScores });
        break;
      }
      case "ActionLiuJu": {
        sendToServer("liuju", {});
        break;
      }
      case "NotifyGameEndResult":
      case "GameEndResult": {
        const result = obj.result || obj;
        const players = result.players || [];
        const scores = players.map(p => p.total_point || p.totalPoint || p.grading?.score || 0);
        const ranks = players.map((_, i) => i + 1);
        const deltaPt = players[0]?.grading?.delta_point || players[0]?.grading?.deltaPoint || 0;
        sendToServer("gameEnd", { scores, ranks, deltaPt });
        break;
      }
    }
  }

  // --- protobuf.js を検出 ---
  function findProtobuf() {
    // protobuf.js の一般的なグローバル名
    const candidates = [
      window.protobuf,
      window.$protobuf,
      window.dcodeIO && window.dcodeIO.protobuf,
    ];
    for (const pb of candidates) {
      if (pb && pb.Type && pb.Type.prototype && pb.Type.prototype.decode) {
        return pb;
      }
    }
    return null;
  }

  // ポーリングで検出 (ゲームのスクリプト読み込み後に protobuf.js が利用可能になる)
  const poller = setInterval(() => {
    const pb = findProtobuf();
    if (pb) {
      clearInterval(poller);
      hookProtobuf(pb);
      connectServer();
    }
  }, 300);
  setTimeout(() => {
    clearInterval(poller);
    clearInterval(poller);
  }, 30000);

  // window.protobuf がセットされた瞬間もキャッチ
  let _pb = window.protobuf;
  try {
    Object.defineProperty(window, "protobuf", {
      get() { return _pb; },
      set(v) {
        _pb = v;
        if (v) hookProtobuf(v);
      },
      configurable: true,
    });
  } catch (e) {}

  let _$pb = window.$protobuf;
  try {
    Object.defineProperty(window, "$protobuf", {
      get() { return _$pb; },
      set(v) {
        _$pb = v;
        if (v) hookProtobuf(v);
      },
      configurable: true,
    });
  } catch (e) {}

})();
