import { fork } from "node:child_process";
import { serve } from "@hono/node-server";
import { Hono } from "hono";

const execPhp = (code, input, timeoutMsec) => {
	return new Promise((resolve, _reject) => {
		const proc = fork("./exec.mjs");

		proc.send({ code, input });

		proc.on("message", (result) => {
			resolve(result);
			proc.kill();
		});

		setTimeout(() => {
			resolve({
				status: "timeout",
				stdout: "",
				stderr: `Time Limit Exceeded: ${timeoutMsec} msec`,
			});
			proc.kill();
		}, timeoutMsec);
	});
};

const app = new Hono();

app.post("/exec", async (c) => {
	console.log("worker/exec");
	const { code, stdin, max_duration_ms } = await c.req.json();
	const result = await execPhp(code, stdin, max_duration_ms);
	return c.json(result);
});

serve({
	fetch: app.fetch,
	port: 80,
});
