import { useAtomValue, useSetAtom } from "jotai";
import { useHydrateAtoms } from "jotai/utils";
import { useContext, useEffect, useState } from "react";
import { useTimer } from "react-use-precision-timer";
import { useDebouncedCallback } from "use-debounce";
import { ApiClientContext } from "../api/client";
import type { components } from "../api/schema";
import {
	gameStateKindAtom,
	handleSubmitCodePostAtom,
	handleSubmitCodePreAtom,
	setCurrentTimestampAtom,
	setDurationSecondsAtom,
	setGameStartedAtAtom,
	setLatestGameStateAtom,
} from "../states/play";
import GolfPlayAppGaming from "./GolfPlayApps/GolfPlayAppGaming";
import GolfPlayAppLoading from "./GolfPlayApps/GolfPlayAppLoading";
import GolfPlayAppStarting from "./GolfPlayApps/GolfPlayAppStarting";
import GolfPlayAppWaiting from "./GolfPlayApps/GolfPlayAppWaiting";

type Game = components["schemas"]["Game"];
type User = components["schemas"]["User"];
type LatestGameState = components["schemas"]["LatestGameState"];

type Props = {
	game: Game;
	player: User;
	initialGameState: LatestGameState;
};

export default function GolfPlayApp({ game, player, initialGameState }: Props) {
	useHydrateAtoms([
		[setDurationSecondsAtom, game.duration_seconds],
		[setGameStartedAtAtom, game.started_at ?? null],
		[setLatestGameStateAtom, initialGameState],
	]);

	const apiClient = useContext(ApiClientContext)!;

	const gameStateKind = useAtomValue(gameStateKindAtom);
	const setGameStartedAt = useSetAtom(setGameStartedAtAtom);
	const setCurrentTimestamp = useSetAtom(setCurrentTimestampAtom);
	const handleSubmitCodePre = useSetAtom(handleSubmitCodePreAtom);
	const handleSubmitCodePost = useSetAtom(handleSubmitCodePostAtom);
	const setLatestGameState = useSetAtom(setLatestGameStateAtom);

	useTimer({ delay: 1000, startImmediately: true }, setCurrentTimestamp);

	const playerProfile = {
		id: player.user_id,
		displayName: player.display_name,
		iconPath: player.icon_path ?? null,
	};

	const onCodeChange = useDebouncedCallback(async (code: string) => {
		if (game.game_type === "1v1") {
			console.log("player:c2s:code");
			await apiClient.postGamePlayCode(game.game_id, code);
		}
	}, 1000);

	const onCodeSubmit = useDebouncedCallback(
		async (code: string) => {
			if (code === "") {
				return;
			}
			console.log("player:c2s:submit");
			handleSubmitCodePre();
			await apiClient.postGamePlaySubmit(game.game_id, code);
			await new Promise((resolve) => setTimeout(resolve, 1000));
			handleSubmitCodePost();
		},
		1000,
		{ leading: true },
	);

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
					const { state } = await apiClient.getGamePlayLatestState(
						game.game_id,
					);
					setLatestGameState(state);
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
		setLatestGameState,
	]);

	if (gameStateKind === "loading") {
		return <GolfPlayAppLoading />;
	} else if (gameStateKind === "waiting") {
		return (
			<GolfPlayAppWaiting
				gameDisplayName={game.display_name}
				playerProfile={playerProfile}
			/>
		);
	} else if (gameStateKind === "starting") {
		return <GolfPlayAppStarting gameDisplayName={game.display_name} />;
	} else if (gameStateKind === "gaming" || gameStateKind === "finished") {
		return (
			<GolfPlayAppGaming
				gameDisplayName={game.display_name}
				playerProfile={playerProfile}
				problemTitle={game.problem.title}
				problemDescription={game.problem.description}
				problemLanguage={game.problem.language}
				sampleCode={game.problem.sample_code}
				initialCode={initialGameState.code}
				onCodeChange={onCodeChange}
				onCodeSubmit={onCodeSubmit}
				isFinished={gameStateKind === "finished"}
			/>
		);
	}
}
