/**
 * @vitest-environment jsdom
 */
import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, test } from "vitest";
import SubmitButton from "./SubmitButton";

afterEach(() => {
	cleanup();
});

describe("SubmitButton", () => {
	test("renders children text", () => {
		render(<SubmitButton>Submit</SubmitButton>);
		expect(screen.getByText("Submit")).toBeDefined();
	});

	test("renders as a button element", () => {
		render(<SubmitButton>Click</SubmitButton>);
		const button = screen.getByText("Click");
		expect(button.tagName).toBe("BUTTON");
	});

	test("can be disabled", () => {
		render(<SubmitButton disabled>Submit</SubmitButton>);
		const button = screen.getByText("Submit") as HTMLButtonElement;
		expect(button.disabled).toBe(true);
	});

	test("is not disabled by default", () => {
		render(<SubmitButton>Submit</SubmitButton>);
		const button = screen.getByText("Submit") as HTMLButtonElement;
		expect(button.disabled).toBe(false);
	});

	test("has sky-600 background styling", () => {
		render(<SubmitButton>Submit</SubmitButton>);
		const button = screen.getByText("Submit");
		expect(button.className).toContain("bg-sky-600");
	});
});
