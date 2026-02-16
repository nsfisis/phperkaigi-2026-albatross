/**
 * @vitest-environment jsdom
 */
import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, test } from "vitest";
import BorderedContainer from "./BorderedContainer";

afterEach(() => {
	cleanup();
});

describe("BorderedContainer", () => {
	test("renders children", () => {
		render(<BorderedContainer>Hello World</BorderedContainer>);
		expect(screen.getByText("Hello World")).toBeDefined();
	});

	test("applies custom className", () => {
		render(
			<BorderedContainer className="custom-class">Content</BorderedContainer>,
		);
		const container = screen.getByText("Content").closest("div");
		expect(container?.className).toContain("custom-class");
	});

	test("has default border styling", () => {
		render(<BorderedContainer>Styled</BorderedContainer>);
		const container = screen.getByText("Styled").closest("div");
		expect(container?.className).toContain("border-2");
		expect(container?.className).toContain("border-blue-600");
		expect(container?.className).toContain("rounded-xl");
	});
});
