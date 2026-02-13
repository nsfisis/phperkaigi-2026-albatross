import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";
import tsconfigPaths from "vite-tsconfig-paths";

export default defineConfig({
	base: process.env.ALBATROSS_BASE_PATH || "/",
	plugins: [tailwindcss(), react(), tsconfigPaths()],
});
