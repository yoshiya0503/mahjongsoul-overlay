import { connect } from "./js/ws.js";
import { render } from "./js/renderer.js";

document.getElementById("session-clear").addEventListener("click", () => {
  fetch("/api/session/clear", { method: "POST" });
});

connect(render);
