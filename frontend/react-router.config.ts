import type { Config } from "@react-router/dev/config";

export default {
	basename: process.env.ALBATROSS_BASE_PATH || "/",
} satisfies Config;
