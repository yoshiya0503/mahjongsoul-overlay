// 設定値・定数
window.__MJS_CONSTANTS = {
  SERVER_URL: "wss://mahjongsoul-overlay.fly.dev/ws/hook",
  PING_INTERVAL: 30000,
  RECONNECT_DELAY: 5000,
  POLL_INTERVAL: 2000,

  // level.id: 上2桁=モード(10=四麻,20=三麻), 次2桁=段位, 最後1桁=星
  RANK_NAMES: {
    1: "初心", 2: "雀士", 3: "雀傑", 4: "雀豪", 5: "雀聖", 6: "魂天",
  },
  RANK_TARGETS: {
    101: 20, 102: 80, 103: 200,
    201: 600, 202: 800, 203: 1000,
    301: 1200, 302: 1400, 303: 2000,
    401: 2800, 402: 3200, 403: 3600,
    501: 4000, 502: 6000, 503: 9000,
    601: 0,
  },
};
