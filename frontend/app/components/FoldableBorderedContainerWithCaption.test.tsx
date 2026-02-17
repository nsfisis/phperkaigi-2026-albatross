/**
 * @vitest-environment jsdom
 */
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, test } from "vitest";
import FoldableBorderedContainerWithCaption from "./FoldableBorderedContainerWithCaption";

afterEach(() => {
	cleanup();
});

describe("FoldableBorderedContainerWithCaption", () => {
	test("renders caption", () => {
		render(
			<FoldableBorderedContainerWithCaption caption="Foldable Title">
				Content
			</FoldableBorderedContainerWithCaption>,
		);
		expect(screen.getByText("Foldable Title")).toBeDefined();
	});

	test("shows children by default (open state)", () => {
		render(
			<FoldableBorderedContainerWithCaption caption="Title">
				<div data-testid="child">Visible</div>
			</FoldableBorderedContainerWithCaption>,
		);
		const child = screen.getByTestId("child");
		expect(child.parentElement?.className).not.toContain("hidden");
	});

	test("hides children when toggle button is clicked", () => {
		render(
			<FoldableBorderedContainerWithCaption caption="Title">
				<div data-testid="child">Content</div>
			</FoldableBorderedContainerWithCaption>,
		);
		const toggleButton = screen.getByRole("button");
		fireEvent.click(toggleButton);
		const child = screen.getByTestId("child");
		expect(child.parentElement?.className).toContain("hidden");
	});

	test("shows children again when toggle button is clicked twice", () => {
		render(
			<FoldableBorderedContainerWithCaption caption="Title">
				<div data-testid="child">Content</div>
			</FoldableBorderedContainerWithCaption>,
		);
		const toggleButton = screen.getByRole("button");
		fireEvent.click(toggleButton);
		fireEvent.click(toggleButton);
		const child = screen.getByTestId("child");
		expect(child.parentElement?.className).not.toContain("hidden");
	});
});
