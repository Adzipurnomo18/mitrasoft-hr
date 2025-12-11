import "./styles/base.css";
import "./styles/components.css";
import "./styles/layout.css";

import App from "./App.svelte";

const app = new App({
  target: document.getElementById("app")
});

export default app;
