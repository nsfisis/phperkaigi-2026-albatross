import PHPWasm from "./php-wasm.js";
import { buildResult, createIOCallbacks, preprocessCode } from "./lib.mjs";

process.once("message", async ({ code: originalCode, input }) => {
	const code = preprocessCode(originalCode);
	const io = createIOCallbacks(input);

	const { ccall } = await PHPWasm({
		stdin: io.stdin,
		stdout: io.stdout,
		stderr: io.stderr,
	});

	let err;
	let result;
	try {
		result = ccall("php_wasm_run", "number", ["string"], [code]);
	} catch (e) {
		err = e;
	}

	process.send(buildResult(err, result, io.getStdout, io.getStderr));
});
