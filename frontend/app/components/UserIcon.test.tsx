/**
 * @vitest-environment jsdom
 */
import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, test } from "vitest";
import UserIcon from "./UserIcon";

afterEach(() => {
	cleanup();
});

describe("UserIcon", () => {
	test("renders an img element", () => {
		render(
			<UserIcon
				iconPath="icons/test.png"
				displayName="TestUser"
				className="w-16 h-16"
			/>,
		);
		const img = screen.getByAltText("TestUser のアイコン");
		expect(img.tagName).toBe("IMG");
	});

	test("sets alt text with display name", () => {
		render(
			<UserIcon
				iconPath="icons/test.png"
				displayName="Alice"
				className="w-16 h-16"
			/>,
		);
		expect(screen.getByAltText("Alice のアイコン")).toBeDefined();
	});

	test("applies rounded-full and border classes", () => {
		render(
			<UserIcon
				iconPath="icons/test.png"
				displayName="Bob"
				className="w-16 h-16"
			/>,
		);
		const img = screen.getByAltText("Bob のアイコン");
		expect(img.className).toContain("rounded-full");
		expect(img.className).toContain("border-4");
		expect(img.className).toContain("border-white");
	});

	test("applies custom className", () => {
		render(
			<UserIcon
				iconPath="icons/test.png"
				displayName="Bob"
				className="w-48 h-48"
			/>,
		);
		const img = screen.getByAltText("Bob のアイコン");
		expect(img.className).toContain("w-48");
		expect(img.className).toContain("h-48");
	});
});
