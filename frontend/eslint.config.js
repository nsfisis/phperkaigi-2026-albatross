import js from "@eslint/js";
import jsxA11y from "eslint-plugin-jsx-a11y";
import react from "eslint-plugin-react";
import reactHooks from "eslint-plugin-react-hooks";
import { defineConfig, globalIgnores } from "eslint/config";
import globals from "globals";
import ts from "typescript-eslint";

export default defineConfig(
	globalIgnores(["node_modules/", "dist/"]),
	js.configs.recommended,
	ts.configs.recommended,
	react.configs.flat.recommended,
	react.configs.flat["jsx-runtime"],
	reactHooks.configs["recommended-latest"],
	jsxA11y.flatConfigs.recommended,
	{
		languageOptions: {
			globals: {
				...globals.browser,
				...globals.node,
			},
		},
	},
	{
		settings: {
			react: {
				version: "detect",
			},
			linkComponents: [{ name: "Link", linkAttribute: "to" }],
		},
	},
);
