/**
 * @vitest-environment jsdom
 */
import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, test } from "vitest";
import BorderedContainerWithCaption from "./BorderedContainerWithCaption";

afterEach(() => {
	cleanup();
});

describe("BorderedContainerWithCaption", () => {
	test("renders caption as heading", () => {
		render(
			<BorderedContainerWithCaption caption="Test Caption">
				Content
			</BorderedContainerWithCaption>,
		);
		expect(screen.getByText("Test Caption")).toBeDefined();
		expect(screen.getByText("Test Caption").tagName).toBe("H2");
	});

	test("renders children", () => {
		render(
			<BorderedContainerWithCaption caption="Title">
				Child Content
			</BorderedContainerWithCaption>,
		);
		expect(screen.getByText("Child Content")).toBeDefined();
	});

	test("wraps in bordered container with blue border", () => {
		render(
			<BorderedContainerWithCaption caption="Title">
				Content
			</BorderedContainerWithCaption>,
		);
		const container = screen.getByText("Content").closest(".border-2");
		expect(container).not.toBeNull();
	});
});
