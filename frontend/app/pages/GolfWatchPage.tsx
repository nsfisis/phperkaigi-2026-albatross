import { createStore, Provider as JotaiProvider } from "jotai";
import { useEffect, useMemo, useState } from "react";
import { useLocation } from "wouter";
import { ApiClientContext, createApiClient } from "../api/client";
import type { components } from "../api/schema";
import GolfWatchApp from "../components/GolfWatchApp";
import { APP_NAME } from "../config";
import { usePageTitle } from "../hooks/usePageTitle";

type Game = components["schemas"]["Game"];
type LatestGameState = components["schemas"]["LatestGameState"];
type RankingEntry = components["schemas"]["RankingEntry"];

export default function GolfWatchPage({ gameId }: { gameId: string }) {
	const [, navigate] = useLocation();

	const [game, setGame] = useState<Game | null>(null);
	const [ranking, setRanking] = useState<RankingEntry[]>([]);
	const [gameStates, setGameStates] = useState<{
		[key: string]: LatestGameState;
	}>({});
	const [loading, setLoading] = useState(true);

	const gameIdNum = Number(gameId);

	usePageTitle(
		game
			? `Golf Watching ${game.display_name} | ${APP_NAME}`
			: `Golf Watching | ${APP_NAME}`,
	);

	useEffect(() => {
		const apiClient = createApiClient();
		Promise.all([
			apiClient.getGame(gameIdNum),
			apiClient.getGameWatchRanking(gameIdNum),
			apiClient.getGameWatchLatestStates(gameIdNum),
		])
			.then(([{ game }, { ranking }, { states }]) => {
				setGame(game);
				setRanking(ranking);
				setGameStates(states);
			})
			.catch(() => navigate("/dashboard"))
			.finally(() => setLoading(false));
	}, [gameIdNum, navigate]);

	const store = useMemo(() => {
		if (!game) return null;
		return createStore();
	}, [game]);

	if (loading || !game || !store) {
		return (
			<div className="min-h-screen bg-gray-100 flex items-center justify-center">
				<p className="text-gray-500">Loading...</p>
			</div>
		);
	}

	return (
		<JotaiProvider store={store}>
			<ApiClientContext.Provider value={createApiClient()}>
				<GolfWatchApp
					key={game.game_id}
					game={game}
					initialGameStates={gameStates}
					initialRanking={ranking}
				/>
			</ApiClientContext.Provider>
		</JotaiProvider>
	);
}
