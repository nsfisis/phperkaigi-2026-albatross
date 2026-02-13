import { Provider as JotaiProvider, createStore } from "jotai";
import { useEffect, useMemo, useState } from "react";
import { useLocation } from "wouter";
import { ApiClientContext, createApiClient } from "../api/client";
import type { components } from "../api/schema";
import { getToken } from "../auth";
import GolfPlayApp from "../components/GolfPlayApp";
import { APP_NAME } from "../config";
import { useAuth } from "../hooks/useAuth";
import { usePageTitle } from "../hooks/usePageTitle";

type Game = components["schemas"]["Game"];
type LatestGameState = components["schemas"]["LatestGameState"];

export default function GolfPlayPage({ gameId }: { gameId: string }) {
	const { user } = useAuth();
	const [, navigate] = useLocation();

	const [game, setGame] = useState<Game | null>(null);
	const [gameState, setGameState] = useState<LatestGameState | null>(null);
	const [loading, setLoading] = useState(true);

	const gameIdNum = Number(gameId);

	usePageTitle(
		game
			? `Golf Playing ${game.display_name} | ${APP_NAME}`
			: `Golf Playing | ${APP_NAME}`,
	);

	useEffect(() => {
		const token = getToken();
		if (!token) return;
		const apiClient = createApiClient(token);
		Promise.all([
			apiClient.getGame(gameIdNum),
			apiClient.getGamePlayLatestState(gameIdNum),
		])
			.then(([{ game }, { state }]) => {
				setGame(game);
				setGameState(state);
			})
			.catch(() => navigate("/dashboard"))
			.finally(() => setLoading(false));
	}, [gameIdNum, navigate]);

	const store = useMemo(() => {
		if (!game || !user) return null;
		return createStore();
	}, [game, user]);

	if (loading || !game || !gameState || !user || !store) {
		return (
			<div className="min-h-screen bg-gray-100 flex items-center justify-center">
				<p className="text-gray-500">Loading...</p>
			</div>
		);
	}

	const token = getToken()!;

	return (
		<JotaiProvider store={store}>
			<ApiClientContext.Provider value={createApiClient(token)}>
				<GolfPlayApp
					key={game.game_id}
					game={game}
					player={user}
					initialGameState={gameState}
				/>
			</ApiClientContext.Provider>
		</JotaiProvider>
	);
}
