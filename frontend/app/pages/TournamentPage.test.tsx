/**
 * @vitest-environment jsdom
 */
import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, test } from "vitest";
import TournamentPage, { standardBracketSeedsForTest } from "./TournamentPage";

afterEach(() => {
	cleanup();
});

describe("standardBracketSeeds", () => {
	test("bracket_size=2 returns [1, 2]", () => {
		const seeds = standardBracketSeedsForTest(2);
		expect(seeds).toEqual([1, 2]);
	});

	test("bracket_size=4 returns [1, 4, 2, 3]", () => {
		const seeds = standardBracketSeedsForTest(4);
		expect(seeds).toEqual([1, 4, 2, 3]);
	});

	test("bracket_size=8 returns [1, 8, 4, 5, 2, 7, 3, 6]", () => {
		const seeds = standardBracketSeedsForTest(8);
		expect(seeds).toEqual([1, 8, 4, 5, 2, 7, 3, 6]);
	});

	test("all seeds present for size 16", () => {
		const seeds = standardBracketSeedsForTest(16);
		expect(seeds).toHaveLength(16);
		const sorted = [...seeds].sort((a, b) => a - b);
		expect(sorted).toEqual(Array.from({ length: 16 }, (_, i) => i + 1));
	});

	test("seed 1 and seed 2 on opposite sides for size 8", () => {
		const seeds = standardBracketSeedsForTest(8);
		const pos1 = seeds.indexOf(1);
		const pos2 = seeds.indexOf(2);
		// Seed 1 in first half (0-3), Seed 2 in second half (4-7)
		expect(pos1).toBeLessThan(4);
		expect(pos2).toBeGreaterThanOrEqual(4);
	});
});

describe("TournamentPage", () => {
	test("shows loading state initially", () => {
		render(<TournamentPage tournamentId="1" />);
		expect(screen.getByText("Loading...")).toBeDefined();
	});

	test("shows error for invalid tournament ID", () => {
		render(<TournamentPage tournamentId="abc" />);
		expect(screen.getByText("Invalid tournament ID")).toBeDefined();
	});

	test("shows error for zero tournament ID", () => {
		render(<TournamentPage tournamentId="0" />);
		expect(screen.getByText("Invalid tournament ID")).toBeDefined();
	});
});
