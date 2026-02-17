import { createStore } from "jotai";
import { describe, expect, test } from "vitest";
import {
	calcCodeSize,
	gameStateKindAtom,
	gamingLeftTimeSecondsAtom,
	handleSubmitCodePostAtom,
	handleSubmitCodePreAtom,
	scoreAtom,
	setCurrentTimestampAtom,
	setDurationSecondsAtom,
	setGameStartedAtAtom,
	setLatestGameStateAtom,
	startingLeftTimeSecondsAtom,
	statusAtom,
} from "./play";

describe("calcCodeSize", () => {
	test("counts UTF-8 bytes after removing whitespace (swift)", () => {
		expect(calcCodeSize("print(1)", "swift")).toBe(8);
	});

	test("removes all whitespace for swift", () => {
		expect(calcCodeSize("print( 1 )\n", "swift")).toBe(8);
	});

	test("removes <?php tag for php", () => {
		expect(calcCodeSize("<?php echo 1;", "php")).toBe(6);
	});

	test("removes <? short tag for php", () => {
		expect(calcCodeSize("<? echo 1;", "php")).toBe(6);
	});

	test("removes ?> closing tag for php", () => {
		expect(calcCodeSize("<?php echo 1;?>", "php")).toBe(6);
	});

	test("removes whitespace and tags together for php", () => {
		expect(calcCodeSize("<?php\n echo 1; \n?>", "php")).toBe(6);
	});

	test("returns 0 for empty string", () => {
		expect(calcCodeSize("", "swift")).toBe(0);
	});

	test("counts multi-byte characters correctly", () => {
		// "あ" is 3 bytes in UTF-8
		expect(calcCodeSize("あ", "swift")).toBe(3);
	});

	test("php with only tags and whitespace returns 0", () => {
		expect(calcCodeSize("<?php ?>", "php")).toBe(0);
	});
});

describe("Jotai atoms", () => {
	test("gameStateKindAtom returns 'loading' initially", () => {
		const store = createStore();
		expect(store.get(gameStateKindAtom)).toBe("loading");
	});

	test("gameStateKindAtom returns 'waiting' when timestamp set but no startedAt", () => {
		const store = createStore();
		store.set(setCurrentTimestampAtom);
		expect(store.get(gameStateKindAtom)).toBe("waiting");
	});

	test("gameStateKindAtom returns 'starting' when now < startedAt", () => {
		const store = createStore();
		const now = Math.floor(Date.now() / 1000);
		store.set(setCurrentTimestampAtom);
		store.set(setGameStartedAtAtom, now + 60);
		store.set(setDurationSecondsAtom, 300);
		expect(store.get(gameStateKindAtom)).toBe("starting");
	});

	test("gameStateKindAtom returns 'gaming' when now >= startedAt and now < finishedAt", () => {
		const store = createStore();
		const now = Math.floor(Date.now() / 1000);
		store.set(setCurrentTimestampAtom);
		store.set(setGameStartedAtAtom, now - 10);
		store.set(setDurationSecondsAtom, 300);
		expect(store.get(gameStateKindAtom)).toBe("gaming");
	});

	test("gameStateKindAtom returns 'finished' when now >= finishedAt", () => {
		const store = createStore();
		const now = Math.floor(Date.now() / 1000);
		store.set(setCurrentTimestampAtom);
		store.set(setGameStartedAtAtom, now - 400);
		store.set(setDurationSecondsAtom, 300);
		expect(store.get(gameStateKindAtom)).toBe("finished");
	});

	test("startingLeftTimeSecondsAtom returns null when startedAt is null", () => {
		const store = createStore();
		expect(store.get(startingLeftTimeSecondsAtom)).toBeNull();
	});

	test("startingLeftTimeSecondsAtom returns null when currentTimestamp is null", () => {
		const store = createStore();
		store.set(setGameStartedAtAtom, 1000);
		expect(store.get(startingLeftTimeSecondsAtom)).toBeNull();
	});

	test("startingLeftTimeSecondsAtom returns remaining time before start", () => {
		const store = createStore();
		const now = Math.floor(Date.now() / 1000);
		store.set(setCurrentTimestampAtom);
		store.set(setGameStartedAtAtom, now + 30);
		const leftTime = store.get(startingLeftTimeSecondsAtom);
		expect(leftTime).toBeGreaterThanOrEqual(29);
		expect(leftTime).toBeLessThanOrEqual(30);
	});

	test("startingLeftTimeSecondsAtom returns 0 when past start time", () => {
		const store = createStore();
		const now = Math.floor(Date.now() / 1000);
		store.set(setCurrentTimestampAtom);
		store.set(setGameStartedAtAtom, now - 10);
		expect(store.get(startingLeftTimeSecondsAtom)).toBe(0);
	});

	test("gamingLeftTimeSecondsAtom returns null when startedAt is null", () => {
		const store = createStore();
		expect(store.get(gamingLeftTimeSecondsAtom)).toBeNull();
	});

	test("gamingLeftTimeSecondsAtom returns remaining game time", () => {
		const store = createStore();
		const now = Math.floor(Date.now() / 1000);
		store.set(setCurrentTimestampAtom);
		store.set(setGameStartedAtAtom, now - 10);
		store.set(setDurationSecondsAtom, 300);
		const leftTime = store.get(gamingLeftTimeSecondsAtom);
		expect(leftTime).toBeGreaterThanOrEqual(289);
		expect(leftTime).toBeLessThanOrEqual(290);
	});

	test("gamingLeftTimeSecondsAtom clamps to 0 when game is over", () => {
		const store = createStore();
		const now = Math.floor(Date.now() / 1000);
		store.set(setCurrentTimestampAtom);
		store.set(setGameStartedAtAtom, now - 400);
		store.set(setDurationSecondsAtom, 300);
		expect(store.get(gamingLeftTimeSecondsAtom)).toBe(0);
	});

	test("gamingLeftTimeSecondsAtom clamps to duration before start", () => {
		const store = createStore();
		const now = Math.floor(Date.now() / 1000);
		store.set(setCurrentTimestampAtom);
		store.set(setGameStartedAtAtom, now + 60);
		store.set(setDurationSecondsAtom, 300);
		expect(store.get(gamingLeftTimeSecondsAtom)).toBe(300);
	});

	test("statusAtom returns 'running' when submitting code", () => {
		const store = createStore();
		store.set(handleSubmitCodePreAtom);
		expect(store.get(statusAtom)).toBe("running");
	});

	test("statusAtom returns raw status when not submitting", () => {
		const store = createStore();
		expect(store.get(statusAtom)).toBe("none");
	});

	test("handleSubmitCodePostAtom resets submitting state", () => {
		const store = createStore();
		store.set(handleSubmitCodePreAtom);
		expect(store.get(statusAtom)).toBe("running");
		store.set(handleSubmitCodePostAtom);
		expect(store.get(statusAtom)).toBe("none");
	});

	test("setLatestGameStateAtom updates status and score", () => {
		const store = createStore();
		store.set(setLatestGameStateAtom, {
			code: "",
			status: "success",
			score: 42,
			best_score_submitted_at: null,
		});
		expect(store.get(statusAtom)).toBe("success");
		expect(store.get(scoreAtom)).toBe(42);
	});

	test("scoreAtom returns null initially", () => {
		const store = createStore();
		expect(store.get(scoreAtom)).toBeNull();
	});
});
