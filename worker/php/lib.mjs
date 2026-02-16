const PRELUDE = `
  define('STDIN', fopen('php://stdin', 'r'));
  define('STDOUT', fopen('php://stdout', 'r'));
  define('STDERR', fopen('php://stderr', 'r'));

  error_reporting(E_ALL & ~E_WARNING & ~E_NOTICE & ~E_DEPRECATED);

  `;

const BUFFER_MAX = 10 * 1024;

export function preprocessCode(originalCode) {
	if (originalCode.startsWith("<?php")) {
		return PRELUDE + originalCode.slice(5);
	}
	if (originalCode.startsWith("<?")) {
		return PRELUDE + originalCode.slice(2);
	}
	return PRELUDE + originalCode;
}

export function createIOCallbacks(input) {
	let stdinPos = 0;
	const stdinBuf = Buffer.from(input);
	let stdoutPos = 0;
	const stdoutBuf = Buffer.alloc(BUFFER_MAX);
	let stderrPos = 0;
	const stderrBuf = Buffer.alloc(BUFFER_MAX);

	return {
		stdin: () => {
			if (stdinBuf.length <= stdinPos) {
				return null;
			}
			return stdinBuf.readUInt8(stdinPos++);
		},
		stdout: (asciiCode) => {
			if (asciiCode === null) {
				return;
			}
			if (BUFFER_MAX <= stdoutPos) {
				return;
			}
			stdoutBuf.writeUInt8(
				asciiCode < 0 ? asciiCode + 256 : asciiCode,
				stdoutPos++,
			);
		},
		stderr: (asciiCode) => {
			if (asciiCode === null) {
				return;
			}
			if (BUFFER_MAX <= stderrPos) {
				return;
			}
			stderrBuf.writeUInt8(
				asciiCode < 0 ? asciiCode + 256 : asciiCode,
				stderrPos++,
			);
		},
		getStdout: () => stdoutBuf.subarray(0, stdoutPos).toString(),
		getStderr: () => stderrBuf.subarray(0, stderrPos).toString(),
	};
}

export function buildResult(err, ccallResult, getStdout, getStderr) {
	if (err) {
		return {
			status: "runtime_error",
			stdout: getStdout(),
			stderr: `${getStderr()}\n${err.toString()}`,
		};
	}
	return {
		status: ccallResult === 0 ? "success" : "runtime_error",
		stdout: getStdout(),
		stderr: getStderr(),
	};
}
