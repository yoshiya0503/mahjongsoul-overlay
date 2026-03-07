const WS_URL = `${location.protocol === "https:" ? "wss:" : "ws:"}//${location.host}/ws/overlay`;

let ws = null;
let reconnectTimer = null;
let pingTimer = null;

export function connect(onMessage) {
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
      onMessage(JSON.parse(event.data));
    } catch (e) {}
  };
  ws.onclose = () => {
    clearInterval(pingTimer);
    reconnectTimer = setTimeout(() => connect(onMessage), 2000);
  };
}
