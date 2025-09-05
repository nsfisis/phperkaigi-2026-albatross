import { useAtomValue, useSetAtom } from "jotai";
import { useHydrateAtoms } from "jotai/utils";
import { useContext, useEffect, useState } from "react";
import { useTimer } from "react-use-precision-timer";
import { ApiClientContext } from "../api/client";
import type { components } from "../api/schema";
import {
	gameStateKindAtom,
	rankingAtom,
	setCurrentTimestampAtom,
	setDurationSecondsAtom,
	setGameStartedAtAtom,
	setLatestGameStatesAtom,
} from "../states/watch";
import GolfWatchAppGaming1v1 from "./GolfWatchApps/GolfWatchAppGaming1v1";
import GolfWatchAppGamingMultiplayer from "./GolfWatchApps/GolfWatchAppGamingMultiplayer";
import GolfWatchAppLoading from "./GolfWatchApps/GolfWatchAppLoading";
import GolfWatchAppStarting from "./GolfWatchApps/GolfWatchAppStarting";
import GolfWatchAppWaiting1v1 from "./GolfWatchApps/GolfWatchAppWaiting1v1";
import GolfWatchAppWaitingMultiplayer from "./GolfWatchApps/GolfWatchAppWaitingMultiplayer";

type Game = components["schemas"]["Game"];
type LatestGameState = components["schemas"]["LatestGameState"];
type RankingEntry = components["schemas"]["RankingEntry"];

export type Props = {
	game: Game;
	initialGameStates: { [key: string]: LatestGameState };
	initialRanking: RankingEntry[];
};

export default function GolfWatchApp({
	game,
	initialGameStates,
	initialRanking,
}: Props) {
	useHydrateAtoms([
		[rankingAtom, initialRanking],
		[setDurationSecondsAtom, game.duration_seconds],
		[setGameStartedAtAtom, game.started_at ?? null],
		[setLatestGameStatesAtom, initialGameStates],
	]);

	const apiClient = useContext(ApiClientContext)!;

	const gameStateKind = useAtomValue(gameStateKindAtom);
	const setGameStartedAt = useSetAtom(setGameStartedAtAtom);
	const setCurrentTimestamp = useSetAtom(setCurrentTimestampAtom);
	const setLatestGameStates = useSetAtom(setLatestGameStatesAtom);
	const setRanking = useSetAtom(rankingAtom);

	useTimer({ delay: 1000, startImmediately: true }, setCurrentTimestamp);

	const playerA = game.main_players[0];
	const playerB = game.main_players[1];

	const playerProfileA = playerA
		? {
				id: playerA.user_id,
				displayName: playerA.display_name,
				iconPath: playerA.icon_path ?? null,
			}
		: null;
	const playerProfileB = playerB
		? {
				id: playerB.user_id,
				displayName: playerB.display_name,
				iconPath: playerB.icon_path ?? null,
			}
		: null;

	const [isDataPolling, setIsDataPolling] = useState(false);

	useEffect(() => {
		if (isDataPolling) {
			return;
		}
		const timerId = setInterval(async () => {
			if (isDataPolling) {
				return;
			}
			setIsDataPolling(true);

			try {
				if (gameStateKind === "waiting") {
					const { game: g } = await apiClient.getGame(game.game_id);
					if (g.started_at != null) {
						setGameStartedAt(g.started_at);
					}
				} else if (gameStateKind === "gaming") {
					const { states } = await apiClient.getGameWatchLatestStates(
						game.game_id,
					);
					setLatestGameStates(states);
					const { ranking } = await apiClient.getGameWatchRanking(game.game_id);
					setRanking(ranking);
				}
			} catch (error) {
				console.error(error);
			} finally {
				setIsDataPolling(false);
			}
		}, 1000);

		return () => {
			clearInterval(timerId);
		};
	}, [
		isDataPolling,
		apiClient,
		game.game_id,
		gameStateKind,
		setGameStartedAt,
		setLatestGameStates,
		setRanking,
	]);

	if (gameStateKind === "loading") {
		return <GolfWatchAppLoading />;
	} else if (gameStateKind === "waiting") {
		return game.game_type === "1v1" ? (
			<GolfWatchAppWaiting1v1
				gameDisplayName={game.display_name}
				playerProfileA={playerProfileA}
				playerProfileB={playerProfileB}
			/>
		) : (
			<GolfWatchAppWaitingMultiplayer gameDisplayName={game.display_name} />
		);
	} else if (gameStateKind === "starting") {
		return <GolfWatchAppStarting gameDisplayName={game.display_name} />;
	} else if (gameStateKind === "gaming" || gameStateKind === "finished") {
		return game.game_type === "1v1" ? (
			<GolfWatchAppGaming1v1
				gameDisplayName={game.display_name}
				playerProfileA={playerProfileA}
				playerProfileB={playerProfileB}
				problemTitle={game.problem.title}
				problemDescription={game.problem.description}
				problemLanguage={game.problem.language}
				sampleCode={game.problem.sample_code}
			/>
		) : (
			<GolfWatchAppGamingMultiplayer
				gameDisplayName={game.display_name}
				problemTitle={game.problem.title}
				problemDescription={game.problem.description}
				problemLanguage={game.problem.language}
				sampleCode={game.problem.sample_code}
			/>
		);
	}
}
