import { describe, expect, test } from "vitest";
import { API_BASE_PATH, APP_NAME, BASE_PATH } from "./config";

describe("config", () => {
	test("BASE_PATH defaults to /", () => {
		expect(BASE_PATH).toBe("/");
	});

	test("API_BASE_PATH is based on BASE_PATH", () => {
		expect(API_BASE_PATH).toBe(`${BASE_PATH}api/`);
	});

	test("APP_NAME is defined", () => {
		expect(APP_NAME).toBeTruthy();
	});
});
