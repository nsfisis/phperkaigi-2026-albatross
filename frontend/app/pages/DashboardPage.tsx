import { useEffect, useState } from "react";
import { useLocation } from "wouter";
import { createApiClient } from "../api/client";
import type { components } from "../api/schema";
import BorderedContainerWithCaption from "../components/BorderedContainerWithCaption";
import NavigateLink from "../components/NavigateLink";
import UserIcon from "../components/UserIcon";
import { APP_NAME, BASE_PATH } from "../config";
import { useAuth } from "../hooks/useAuth";
import { usePageTitle } from "../hooks/usePageTitle";

type Game = components["schemas"]["Game"];

export default function DashboardPage() {
	usePageTitle(`Dashboard | ${APP_NAME}`);

	const { user, logout } = useAuth();
	const [, navigate] = useLocation();

	const [games, setGames] = useState<Game[]>([]);
	const [loading, setLoading] = useState(true);

	useEffect(() => {
		const apiClient = createApiClient();
		apiClient
			.getGames()
			.then(({ games }) => setGames(games))
			.finally(() => setLoading(false));
	}, []);

	async function handleLogout() {
		await logout();
		navigate("/");
	}

	if (loading) {
		return (
			<div className="min-h-screen bg-gray-100 flex items-center justify-center">
				<p className="text-gray-500">Loading...</p>
			</div>
		);
	}

	return (
		<div className="p-6 bg-gray-100 min-h-screen flex flex-col items-center gap-4">
			{user?.icon_path && (
				<UserIcon
					iconPath={user.icon_path}
					displayName={user.display_name}
					className="w-24 h-24"
				/>
			)}
			<h1 className="text-3xl font-bold text-gray-800">{user?.display_name}</h1>
			<BorderedContainerWithCaption caption="試合一覧">
				<div className="px-4">
					{games.length === 0 ? (
						<p>エントリーできる試合はありません</p>
					) : (
						<ul className="divide-y divide-gray-300">
							{games.map((game) => (
								<li
									key={game.game_id}
									className="flex justify-between items-center py-2 gap-4"
								>
									<div>
										<span className="font-medium text-gray-800">
											{game.display_name}
										</span>
									</div>
									<div className="flex gap-2">
										<NavigateLink to={`/golf/${game.game_id}/play`}>
											対戦
										</NavigateLink>
										<NavigateLink to={`/golf/${game.game_id}/watch`}>
											観戦
										</NavigateLink>
										<NavigateLink to={`/golf/${game.game_id}/submissions`}>
											提出履歴
										</NavigateLink>
									</div>
								</li>
							))}
						</ul>
					)}
				</div>
			</BorderedContainerWithCaption>
			<button
				type="button"
				onClick={handleLogout}
				className="px-4 py-2 bg-red-500 text-white rounded-sm transition duration-300 hover:bg-red-700 focus:ring-3 focus:ring-red-400 focus:outline-hidden"
			>
				ログアウト
			</button>
			{user?.is_admin && (
				<a
					href={
						import.meta.env.DEV
							? `http://localhost:8004${BASE_PATH}admin/dashboard`
							: `${BASE_PATH}admin/dashboard`
					}
					className="text-lg text-white bg-sky-600 px-4 py-2 rounded-sm transition duration-300 hover:bg-sky-500 focus:ring-3 focus:ring-sky-400 focus:outline-hidden"
				>
					Admin Dashboard
				</a>
			)}
		</div>
	);
}
