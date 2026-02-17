/**
 * @vitest-environment jsdom
 */
import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, test } from "vitest";
import SubmitStatusLabel from "./SubmitStatusLabel";

afterEach(() => {
	cleanup();
});

describe("SubmitStatusLabel", () => {
	test("renders '提出待ち' for none status", () => {
		render(<SubmitStatusLabel status="none" />);
		expect(screen.getByText("提出待ち")).toBeDefined();
	});

	test("renders '実行中...' for running status", () => {
		render(<SubmitStatusLabel status="running" />);
		expect(screen.getByText("実行中...")).toBeDefined();
	});

	test("renders '成功' for success status", () => {
		render(<SubmitStatusLabel status="success" />);
		expect(screen.getByText("成功")).toBeDefined();
	});

	test("renders 'テスト失敗' for wrong_answer status", () => {
		render(<SubmitStatusLabel status="wrong_answer" />);
		expect(screen.getByText("テスト失敗")).toBeDefined();
	});

	test("renders '時間切れ' for timeout status", () => {
		render(<SubmitStatusLabel status="timeout" />);
		expect(screen.getByText("時間切れ")).toBeDefined();
	});

	test("renders 'コンパイルエラー' for compile_error status", () => {
		render(<SubmitStatusLabel status="compile_error" />);
		expect(screen.getByText("コンパイルエラー")).toBeDefined();
	});

	test("renders '実行時エラー' for runtime_error status", () => {
		render(<SubmitStatusLabel status="runtime_error" />);
		expect(screen.getByText("実行時エラー")).toBeDefined();
	});

	test("renders '！内部エラー！' for internal_error status", () => {
		render(<SubmitStatusLabel status="internal_error" />);
		expect(screen.getByText("！内部エラー！")).toBeDefined();
	});
});
