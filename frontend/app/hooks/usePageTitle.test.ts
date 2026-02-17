/**
 * @vitest-environment jsdom
 */
import { renderHook } from "@testing-library/react";
import { describe, expect, test } from "vitest";
import { usePageTitle } from "./usePageTitle";

describe("usePageTitle", () => {
	test("sets document title", () => {
		renderHook(() => usePageTitle("Test Page"));
		expect(document.title).toBe("Test Page");
	});

	test("updates document title when value changes", () => {
		const { rerender } = renderHook(({ title }) => usePageTitle(title), {
			initialProps: { title: "First" },
		});
		expect(document.title).toBe("First");

		rerender({ title: "Second" });
		expect(document.title).toBe("Second");
	});
});
