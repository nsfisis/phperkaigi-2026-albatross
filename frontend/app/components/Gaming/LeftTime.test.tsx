/**
 * @vitest-environment jsdom
 */
import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, test } from "vitest";
import LeftTime from "./LeftTime";

afterEach(() => {
	cleanup();
});

describe("LeftTime", () => {
	test("renders MM:SS format for short durations", () => {
		render(<LeftTime sec={65} />);
		expect(screen.getByText("01:05")).toBeDefined();
	});

	test("renders 00:00 for zero seconds", () => {
		render(<LeftTime sec={0} />);
		expect(screen.getByText("00:00")).toBeDefined();
	});

	test("renders MM:SS with leading zeros", () => {
		render(<LeftTime sec={5} />);
		expect(screen.getByText("00:05")).toBeDefined();
	});

	test("renders 59:59 for max MM:SS range", () => {
		render(<LeftTime sec={3599} />);
		expect(screen.getByText("59:59")).toBeDefined();
	});

	test("renders long format with hours", () => {
		render(<LeftTime sec={3661} />);
		expect(screen.getByText("1h 1m 1s")).toBeDefined();
	});

	test("renders long format with days", () => {
		render(<LeftTime sec={90061} />);
		expect(screen.getByText("1d 1h 1m 1s")).toBeDefined();
	});

	test("renders long format omitting zero day and minute", () => {
		render(<LeftTime sec={3605} />);
		expect(screen.getByText("1h 5s")).toBeDefined();
	});
});
