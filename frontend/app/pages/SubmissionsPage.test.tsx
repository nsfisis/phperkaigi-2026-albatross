/**
 * @vitest-environment jsdom
 */
import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, test } from "vitest";
import SubmissionsPage from "./SubmissionsPage";

afterEach(() => {
	cleanup();
});

describe("SubmissionsPage", () => {
	test("shows loading state initially", () => {
		render(<SubmissionsPage gameId="1" />);
		expect(screen.getByText("Loading...")).toBeDefined();
	});
});
