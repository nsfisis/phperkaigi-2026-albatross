import { toJsxRuntime } from "hast-util-to-jsx-runtime";
import { Fragment, type JSX } from "react";
import { jsx, jsxs } from "react/jsx-runtime";
import { type BundledLanguage, codeToHast } from "./shiki.bundle";

export type { BundledLanguage };

// https://shiki.matsu.io/packages/next
export async function highlight(code: string, lang: BundledLanguage) {
	const out = await codeToHast(code.trimEnd(), {
		lang: lang === "swift" ? "text" : lang,
		theme: "github-light",
	});

	return toJsxRuntime(out, {
		Fragment,
		jsx,
		jsxs,
	}) as JSX.Element;
}
