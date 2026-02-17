/**
 * @vitest-environment jsdom
 */
import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, test } from "vitest";
import InputText from "./InputText";

afterEach(() => {
	cleanup();
});

describe("InputText", () => {
	test("renders an input element", () => {
		render(<InputText data-testid="input" />);
		const input = screen.getByTestId("input");
		expect(input.tagName).toBe("INPUT");
	});

	test("passes placeholder prop", () => {
		render(<InputText placeholder="Enter text" data-testid="input" />);
		const input = screen.getByTestId("input") as HTMLInputElement;
		expect(input.placeholder).toBe("Enter text");
	});

	test("passes type prop", () => {
		render(<InputText type="password" data-testid="input" />);
		const input = screen.getByTestId("input") as HTMLInputElement;
		expect(input.type).toBe("password");
	});

	test("has border styling", () => {
		render(<InputText data-testid="input" />);
		const input = screen.getByTestId("input");
		expect(input.className).toContain("border-sky-600");
	});
});
