import { config } from "@fortawesome/fontawesome-svg-core";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "@fortawesome/fontawesome-svg-core/styles.css";
import "./tailwind.css";
import "./shiki.css";
import App from "./App";

config.autoAddCss = false;

const root = document.getElementById("root");
if (!root) {
	throw new Error("Root element not found");
}

createRoot(root).render(
	<StrictMode>
		<App />
	</StrictMode>,
);
