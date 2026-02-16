import { describe, expect, it } from "vitest";
import { buildResult, createIOCallbacks, preprocessCode } from "./lib.mjs";

describe("preprocessCode", () => {
	it("removes <?php tag and prepends PRELUDE", () => {
		const result = preprocessCode('<?php echo "hello";');
		expect(result).toContain('echo "hello";');
		expect(result).toContain("error_reporting");
		expect(result).not.toContain("<?php");
	});

	it("removes <? short tag and prepends PRELUDE", () => {
		const result = preprocessCode('<? echo "hello";');
		expect(result).toContain('echo "hello";');
		expect(result).toContain("error_reporting");
		expect(result).not.toContain("<?");
	});

	it("prepends PRELUDE when no php tag present", () => {
		const result = preprocessCode('echo "hello";');
		expect(result).toContain('echo "hello";');
		expect(result).toContain("error_reporting");
	});

	it("handles empty string", () => {
		const result = preprocessCode("");
		expect(result).toContain("error_reporting");
	});

	it("does not remove <?php when not at the start", () => {
		const result = preprocessCode('echo "x"; <?php echo "y";');
		expect(result).toContain("<?php");
	});
});

describe("createIOCallbacks", () => {
	describe("stdin", () => {
		it("reads input byte by byte", () => {
			const io = createIOCallbacks("AB");
			expect(io.stdin()).toBe(65); // 'A'
			expect(io.stdin()).toBe(66); // 'B'
		});

		it("returns null at EOF", () => {
			const io = createIOCallbacks("A");
			io.stdin(); // consume 'A'
			expect(io.stdin()).toBeNull();
			expect(io.stdin()).toBeNull();
		});

		it("returns null immediately for empty input", () => {
			const io = createIOCallbacks("");
			expect(io.stdin()).toBeNull();
		});
	});

	describe("stdout", () => {
		it("captures ASCII writes", () => {
			const io = createIOCallbacks("");
			io.stdout(72); // 'H'
			io.stdout(105); // 'i'
			expect(io.getStdout()).toBe("Hi");
		});

		it("ignores null (flush)", () => {
			const io = createIOCallbacks("");
			io.stdout(65);
			io.stdout(null);
			io.stdout(66);
			expect(io.getStdout()).toBe("AB");
		});

		it("corrects negative asciiCode by adding 256", () => {
			const io = createIOCallbacks("");
			// -191 + 256 = 65 = 'A'
			io.stdout(-191);
			expect(io.getStdout()).toBe("A");
		});

		it("truncates output at 10KB buffer limit", () => {
			const io = createIOCallbacks("");
			const limit = 10 * 1024;
			for (let i = 0; i < limit + 100; i++) {
				io.stdout(65);
			}
			expect(io.getStdout().length).toBe(limit);
		});
	});

	describe("stderr", () => {
		it("captures ASCII writes", () => {
			const io = createIOCallbacks("");
			io.stderr(69); // 'E'
			io.stderr(114); // 'r'
			expect(io.getStderr()).toBe("Er");
		});

		it("ignores null (flush)", () => {
			const io = createIOCallbacks("");
			io.stderr(65);
			io.stderr(null);
			expect(io.getStderr()).toBe("A");
		});

		it("corrects negative asciiCode by adding 256", () => {
			const io = createIOCallbacks("");
			// -156 + 256 = 100 = 'd'
			io.stderr(-156);
			expect(io.getStderr()).toBe("d");
		});

		it("truncates output at 10KB buffer limit", () => {
			const io = createIOCallbacks("");
			const limit = 10 * 1024;
			for (let i = 0; i < limit + 100; i++) {
				io.stderr(65);
			}
			expect(io.getStderr().length).toBe(limit);
		});
	});
});

describe("buildResult", () => {
	it("returns success when err is null and result is 0", () => {
		const result = buildResult(null, 0, () => "out", () => "");
		expect(result).toEqual({
			status: "success",
			stdout: "out",
			stderr: "",
		});
	});

	it("returns runtime_error when result is non-zero", () => {
		const result = buildResult(null, 1, () => "out", () => "err");
		expect(result).toEqual({
			status: "runtime_error",
			stdout: "out",
			stderr: "err",
		});
	});

	it("returns runtime_error with concatenated stderr when err is thrown", () => {
		const err = new Error("fatal");
		const result = buildResult(err, undefined, () => "out", () => "err");
		expect(result.status).toBe("runtime_error");
		expect(result.stdout).toBe("out");
		expect(result.stderr).toContain("err");
		expect(result.stderr).toContain("Error: fatal");
	});
});
