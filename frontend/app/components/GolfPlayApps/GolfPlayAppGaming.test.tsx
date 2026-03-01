/**
 * @vitest-environment jsdom
 */
import { cleanup, render, screen } from "@testing-library/react";
import { createStore, Provider } from "jotai";
import { afterEach, describe, expect, test } from "vitest";
import {
	setCurrentTimestampAtom,
	setDurationSecondsAtom,
	setGameStartedAtAtom,
	setLatestGameStateAtom,
} from "../../states/play";
import GolfPlayAppGaming from "./GolfPlayAppGaming";

afterEach(() => {
	cleanup();
});

function createTestStore() {
	const store = createStore();
	const now = Math.floor(Date.now() / 1000);
	store.set(setCurrentTimestampAtom);
	store.set(setDurationSecondsAtom, 600);
	store.set(setGameStartedAtAtom, now - 60);
	store.set(setLatestGameStateAtom, {
		status: "none",
		code: "",
		score: null,
		best_score_submitted_at: null,
	});
	return store;
}

const defaultProps = {
	gameDisplayName: "Test Game",
	playerProfile: {
		id: 1,
		displayName: "Test Player",
		iconPath: null,
	},
	problemTitle: "Test Problem",
	problemDescription: "Description",
	problemLanguage: "php" as const,
	sampleCode: "<?php echo 1;",
	initialCode: "",
	onCodeChange: () => {},
	onCodeSubmit: () => {},
	isFinished: false,
};

describe("GolfPlayAppGaming submission history", () => {
	test("shows placeholder row when no submissions", () => {
		const store = createTestStore();
		render(
			<Provider store={store}>
				<GolfPlayAppGaming {...defaultProps} submissions={[]} />
			</Provider>,
		);
		expect(screen.getByText("提出待ち")).toBeDefined();
		const dashes = screen.getAllByText("-");
		expect(dashes.length).toBe(3);
	});

	test("renders submission rows with status and code size", () => {
		const store = createTestStore();
		const submissions = [
			{
				submission_id: 1,
				game_id: 1,
				status: "success" as const,
				code: "<?php echo 1;",
				code_size: 7,
				created_at: 1740000000,
			},
			{
				submission_id: 2,
				game_id: 1,
				status: "wrong_answer" as const,
				code: "<?php echo 2;",
				code_size: 10,
				created_at: 1740000060,
			},
		];
		render(
			<Provider store={store}>
				<GolfPlayAppGaming {...defaultProps} submissions={submissions} />
			</Provider>,
		);
		expect(screen.getByText("成功")).toBeDefined();
		expect(screen.getByText("テスト失敗")).toBeDefined();
		expect(screen.getByText("7")).toBeDefined();
		expect(screen.getByText("10")).toBeDefined();
	});

	test("renders table headers", () => {
		const store = createTestStore();
		render(
			<Provider store={store}>
				<GolfPlayAppGaming {...defaultProps} submissions={[]} />
			</Provider>,
		);
		expect(screen.getByText("ステータス")).toBeDefined();
		expect(screen.getByText("スコア")).toBeDefined();
		expect(screen.getByText("提出時刻")).toBeDefined();
		expect(screen.getByText("コード")).toBeDefined();
	});
});
