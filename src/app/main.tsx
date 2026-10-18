import { isMyGo, runtime } from "mygo-runtime";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { App } from "./app";
import "./styles/index.css";

// Only macOS runs the content under the title bar (traffic lights over the
// sidebar); styles key off this.
if (isMyGo() && runtime().platform === "darwin") document.documentElement.classList.add("mac");

const root = document.getElementById("root");
if (root) {
  createRoot(root).render(
    <StrictMode>
      <App />
    </StrictMode>,
  );
}
