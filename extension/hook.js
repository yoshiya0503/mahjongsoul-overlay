// 雀魂 WebSocket Hook - ゲーム内部のprotobuf.jsデコード結果をフックする
(function () {
  "use strict";

  if (window.__mjsHookInstalled) return;
  window.__mjsHookInstalled = true;

  const C = window.__MJS_CONSTANTS;

  let hookSocket = null;
  const OrigWebSocket = window.WebSocket;

  // --- サーバー接続 ---
  let pingTimer = null;

  function connectServer() {
    if (hookSocket && hookSocket.readyState <= 1) return;
    try {
      hookSocket = new OrigWebSocket(C.SERVER_URL);
      hookSocket.onopen = () => {
        clearInterval(pingTimer);
        pingTimer = setInterval(() => {
          if (hookSocket && hookSocket.readyState === 1) {
            hookSocket.send(JSON.stringify({ type: "ping" }));
          }
        }, C.PING_INTERVAL);
      };
      hookSocket.onclose = () => {
        clearInterval(pingTimer);
        setTimeout(connectServer, C.RECONNECT_DELAY);
      };
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

  // --- 段位ヘルパー ---
  function getRankName(levelId) {
    const rank = Math.floor((levelId % 10000) / 100);
    const star = levelId % 100;
    return `${C.RANK_NAMES[rank] || "?"}${star}`;
  }

  function getRankTarget(levelId) {
    return C.RANK_TARGETS[levelId % 10000] || 0;
  }

  // --- protobuf.js デコードフック ---
  let myAccountId = null;

  function parseScores(obj) {
    return {
      scores: obj.scores || [],
      deltaScores: obj.delta_scores || obj.deltaScores || [],
    };
  }

  const eventHandlers = {
    ResLogin(obj) {
      if (obj.account_id) myAccountId = obj.account_id;
    },
    ResOauth2Login(obj) {
      if (obj.account_id) myAccountId = obj.account_id;
    },
    ResAuthGame: handleAuthGame,
    ResEnterGame: handleAuthGame,
    ActionNewRound(obj) {
      sendToServer("newRound", {
        chang: obj.chang || 0,
        ju: obj.ju || 0,
        ben: obj.ben || 0,
        scores: obj.scores || [],
      });
    },
    ActionHule(obj) {
      const { scores, deltaScores } = parseScores(obj);
      let winner = -1;
      if (obj.hules && obj.hules.length > 0) {
        winner = obj.hules[0].seat || 0;
      }
      sendToServer("hule", { scores, deltaScores, winner });
    },
    ActionNoTile(obj) {
      const { scores, deltaScores } = parseScores(obj);
      sendToServer("noTile", { scores, deltaScores });
    },
    ActionLiuJu() {
      sendToServer("liuju", {});
    },
    NotifyGameEndResult: handleGameEnd,
    GameEndResult: handleGameEnd,
  };

  function handleAuthGame(obj) {
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

    const me = myAccountId
      ? allPlayers.find(p => p.account_id === myAccountId)
      : allPlayers[0];
    if (me && me.level) {
      sendToServer("rankPoint", {
        currentPt: me.level.score || 0,
        targetPt: getRankTarget(me.level.id),
        rankName: getRankName(me.level.id),
      });
    }
  }

  function handleGameEnd(obj) {
    const result = obj.result || obj;
    const players = result.players || [];
    const scores = new Array(players.length).fill(0);
    let deltaPt = 0;

    // seat → total_point のマップを作る
    for (const p of players) {
      scores[p.seat || 0] = p.total_point || p.totalPoint || 0;
    }

    // 自分のプレイヤーを特定
    const me = myAccountId
      ? players.find(p => p.account_id === myAccountId)
      : null;

    // 順位はスコアの降順で計算（同スコアはseat順）
    let myRank = 1;
    if (me) {
      const mySeat = me.seat || 0;
      const myScore = scores[mySeat];
      for (let i = 0; i < scores.length; i++) {
        if (i !== mySeat && (scores[i] > myScore || (scores[i] === myScore && i < mySeat))) {
          myRank++;
        }
      }
      deltaPt = me.grading?.delta_point || me.grading?.deltaPoint || 0;
    }

    sendToServer("gameEnd", { scores, ranks: [myRank], deltaPt });
  }

  const WATCHED = new Set(Object.keys(eventHandlers));

  function makeDecodeWrapper(type, decodeFn) {
    const wrapper = function (reader, length) {
      const msg = decodeFn.call(this, reader, length);
      try {
        if (WATCHED.has(type.name)) {
          const obj = type.toObject(msg, { defaults: true, longs: Number, enums: Number });
          eventHandlers[type.name](obj);
        }
      } catch (e) {}
      return msg;
    };
    wrapper.__mjsHooked = true;
    return wrapper;
  }

  function hookProtobuf(pb) {
    if (!pb || !pb.Type || !pb.Type.prototype || !pb.Type.prototype.decode) return;
    if (pb.Type.prototype.decode.__mjsHooked) return;

    const origDecode = pb.Type.prototype.decode;
    pb.Type.prototype.decode = function (reader, length) {
      const msg = origDecode.call(this, reader, length);
      const desc = Object.getOwnPropertyDescriptor(this, "decode");
      if (desc && desc.value && !desc.value.__mjsHooked) {
        this.decode = makeDecodeWrapper(this, desc.value);
      }
      try {
        if (WATCHED.has(this.name)) {
          const obj = this.toObject(msg, { defaults: true, longs: Number, enums: Number });
          eventHandlers[this.name](obj);
        }
      } catch (e) {}
      return msg;
    };
    pb.Type.prototype.decode.__mjsHooked = true;
  }

  // --- protobuf.js を検出 ---
  function findProtobuf() {
    const candidates = [
      window.protobuf,
      window.$protobuf,
      window.dcodeIO && window.dcodeIO.protobuf,
    ];
    return candidates.find(
      (pb) => pb && pb.Type && pb.Type.prototype && pb.Type.prototype.decode
    ) || null;
  }

  function watchGlobal(prop) {
    let current = window[prop];
    try {
      Object.defineProperty(window, prop, {
        get() { return current; },
        set(v) {
          current = v;
          if (v) hookProtobuf(v);
        },
        configurable: true,
      });
    } catch (e) {}
  }

  // ポーリングで検出・再フック (protobuf.jsがリロードされても再適用)
  let serverConnected = false;
  setInterval(() => {
    const pb = findProtobuf();
    if (pb) {
      hookProtobuf(pb);
      if (!serverConnected) {
        connectServer();
        serverConnected = true;
      }
    }
  }, C.POLL_INTERVAL);

  // window.protobuf がセットされた瞬間もキャッチ
  watchGlobal("protobuf");
  watchGlobal("$protobuf");

})();
