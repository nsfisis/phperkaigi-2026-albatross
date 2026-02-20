import { useEffect, useState } from "react";
import { useLocation } from "wouter";
import { createApiClient } from "../api/client";
import type { components } from "../api/schema";
import ProblemColumnContent from "../components/Gaming/ProblemColumnContent";
import NavigateLink from "../components/NavigateLink";
import { APP_NAME } from "../config";
import { usePageTitle } from "../hooks/usePageTitle";

type Game = components["schemas"]["Game"];

export default function GolfProblemPreviewPage({ gameId }: { gameId: string }) {
	const [, navigate] = useLocation();
	const [game, setGame] = useState<Game | null>(null);
	const [loading, setLoading] = useState(true);

	const gameIdNum = Number(gameId);

	usePageTitle(
		game
			? `${game.display_name} - 問題プレビュー | ${APP_NAME}`
			: `問題プレビュー | ${APP_NAME}`,
	);

	useEffect(() => {
		const apiClient = createApiClient();
		apiClient
			.getGame(gameIdNum)
			.then(({ game }) => setGame(game))
			.catch(() => navigate("/dashboard"))
			.finally(() => setLoading(false));
	}, [gameIdNum, navigate]);

	if (loading || !game) {
		return (
			<div className="min-h-screen bg-gray-100 flex items-center justify-center">
				<p className="text-gray-500">Loading...</p>
			</div>
		);
	}

	return (
		<div className="p-6 bg-gray-100 min-h-screen flex flex-col items-center gap-4">
			<h1 className="text-3xl font-bold text-gray-800">{game.display_name}</h1>
			<div className="w-full max-w-3xl flex flex-col gap-4">
				<ProblemColumnContent
					description={game.problem.description}
					language={game.problem.language}
					sampleCode={game.problem.sample_code}
				/>
			</div>
			{game.started_at != null && (
				<NavigateLink to={`/golf/${game.game_id}/play`}>
					対戦ページへ
				</NavigateLink>
			)}
			<NavigateLink to="/dashboard">ダッシュボードへ戻る</NavigateLink>
		</div>
	);
}
