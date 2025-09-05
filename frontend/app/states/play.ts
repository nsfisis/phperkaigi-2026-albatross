import { atom } from "jotai";
import type { components } from "../api/schema";
import type { SupportedLanguage } from "../types/SupportedLanguage";

const gameStartedAtAtom = atom<number | null>(null);
export const setGameStartedAtAtom = atom(null, (_, set, value: number | null) =>
	set(gameStartedAtAtom, value),
);

export type GameStateKind =
	| "loading"
	| "waiting"
	| "starting"
	| "gaming"
	| "finished";
type ExecutionStatus = components["schemas"]["ExecutionStatus"];
type LatestGameState = components["schemas"]["LatestGameState"];

export const gameStateKindAtom = atom<GameStateKind>((get) => {
	const now = get(currentTimestampAtom);
	if (!now) {
		return "loading";
	}
	const startedAt = get(gameStartedAtAtom);
	if (!startedAt) {
		return "waiting";
	}
	const durationSeconds = get(durationSecondsAtom);
	const finishedAt = startedAt + durationSeconds;
	if (now < startedAt) {
		return "starting";
	} else if (now < finishedAt) {
		return "gaming";
	} else {
		return "finished";
	}
});

const currentTimestampAtom = atom<number | null>(null);
export const setCurrentTimestampAtom = atom(null, (_, set) =>
	set(currentTimestampAtom, Math.floor(Date.now() / 1000)),
);

const durationSecondsAtom = atom<number>(0);
export const setDurationSecondsAtom = atom(null, (_, set, value: number) =>
	set(durationSecondsAtom, value),
);

export const startingLeftTimeSecondsAtom = atom<number | null>((get) => {
	const startedAt = get(gameStartedAtAtom);
	if (startedAt === null) {
		return null;
	}
	const currentTimestamp = get(currentTimestampAtom);
	if (currentTimestamp === null) {
		return null;
	}
	return Math.max(0, startedAt - currentTimestamp);
});

export const gamingLeftTimeSecondsAtom = atom<number | null>((get) => {
	const startedAt = get(gameStartedAtAtom);
	if (startedAt === null) {
		return null;
	}
	const durationSeconds = get(durationSecondsAtom);
	const finishedAt = startedAt + durationSeconds;
	const currentTimestamp = get(currentTimestampAtom);
	if (currentTimestamp === null) {
		return null;
	}
	return Math.min(durationSeconds, Math.max(0, finishedAt - currentTimestamp));
});

const rawStatusAtom = atom<ExecutionStatus>("none");
const rawScoreAtom = atom<number | null>(null);
export const statusAtom = atom<ExecutionStatus>((get) => {
	const isSubmittingCode = get(isSubmittingCodeAtom);
	if (isSubmittingCode) {
		return "running";
	} else {
		return get(rawStatusAtom);
	}
});
export const scoreAtom = atom<number | null>((get) => {
	return get(rawScoreAtom);
});

const isSubmittingCodeAtom = atom(false);
export const handleSubmitCodePreAtom = atom(null, (_, set) => {
	set(isSubmittingCodeAtom, true);
});
export const handleSubmitCodePostAtom = atom(null, (_, set) => {
	set(isSubmittingCodeAtom, false);
});

export const setLatestGameStateAtom = atom(
	null,
	(_, set, value: LatestGameState) => {
		set(rawStatusAtom, value.status);
		set(rawScoreAtom, value.score);
	},
);

function cleanCode(code: string, language: SupportedLanguage) {
	if (language === "php") {
		return code
			.replace(/\s+/g, "")
			.replace(/^<\?php/, "")
			.replace(/^<\?/, "")
			.replace(/\?>$/, "");
	} else {
		return code.replace(/\s+/g, "");
	}
}

export function calcCodeSize(
	code: string,
	language: SupportedLanguage,
): number {
	const trimmed = cleanCode(code, language);
	const utf8Encoded = new TextEncoder().encode(trimmed);
	return utf8Encoded.length;
}
