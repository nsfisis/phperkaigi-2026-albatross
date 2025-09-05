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
type LatestGameState = components["schemas"]["LatestGameState"];
type RankingEntry = components["schemas"]["RankingEntry"];

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

export const rankingAtom = atom<RankingEntry[]>([]);

const rawLatestGameStatesAtom = atom<{
	[key: string]: LatestGameState | undefined;
}>({});
export const latestGameStatesAtom = atom((get) => get(rawLatestGameStatesAtom));
export const setLatestGameStatesAtom = atom(
	null,
	(_, set, value: { [key: string]: LatestGameState | undefined }) => {
		set(rawLatestGameStatesAtom, value);
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

export type GameResultKind = "winA" | "winB" | "draw";

export function checkGameResultKind(
	gameStateKind: GameStateKind,
	stateA: LatestGameState | null,
	stateB: LatestGameState | null,
): GameResultKind | null {
	if (gameStateKind !== "finished") {
		return null;
	}

	const scoreA = stateA?.score;
	const scoreB = stateB?.score;
	if (scoreA == null && scoreB == null) {
		return "draw";
	}
	if (scoreA == null) {
		return "winB";
	}
	if (scoreB == null) {
		return "winA";
	}
	if (scoreA === scoreB) {
		// If score is non-null, state and best_score_submitted_at should also be non-null.
		const submittedAtA = stateA!.best_score_submitted_at!;
		const submittedAtB = stateB!.best_score_submitted_at!;
		return submittedAtA < submittedAtB ? "winA" : "winB";
	} else {
		return scoreA < scoreB ? "winA" : "winB";
	}
}
