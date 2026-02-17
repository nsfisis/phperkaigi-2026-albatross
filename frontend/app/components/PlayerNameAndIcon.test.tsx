/**
 * @vitest-environment jsdom
 */
import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, test } from "vitest";
import PlayerNameAndIcon from "./PlayerNameAndIcon";

afterEach(() => {
	cleanup();
});

describe("PlayerNameAndIcon", () => {
	test("renders display name", () => {
		render(
			<PlayerNameAndIcon
				profile={{ id: 1, displayName: "Alice", iconPath: null }}
			/>,
		);
		expect(screen.getByText("Alice")).toBeDefined();
	});

	test("does not render icon when iconPath is null", () => {
		render(
			<PlayerNameAndIcon
				profile={{ id: 1, displayName: "Bob", iconPath: null }}
			/>,
		);
		expect(screen.queryByRole("img")).toBeNull();
	});

	test("renders icon when iconPath is provided", () => {
		render(
			<PlayerNameAndIcon
				profile={{ id: 1, displayName: "Carol", iconPath: "icons/carol.png" }}
			/>,
		);
		expect(screen.getByAltText("Carol のアイコン")).toBeDefined();
	});
});
