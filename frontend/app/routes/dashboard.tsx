import type { LoaderFunctionArgs, MetaFunction } from "react-router";
import { Form, useLoaderData } from "react-router";
import { ensureUserLoggedIn } from "../.server/auth";
import { createApiClient } from "../api/client";
import BorderedContainerWithCaption from "../components/BorderedContainerWithCaption";
import NavigateLink from "../components/NavigateLink";
import UserIcon from "../components/UserIcon";
import { BASE_PATH } from "../config";

export const meta: MetaFunction = () => [
	{ title: "Dashboard | PHPerKaigi 2025 Albatross" },
];

export async function loader({ request }: LoaderFunctionArgs) {
	const { user, token } = await ensureUserLoggedIn(request);
	const apiClient = createApiClient(token);

	const { games } = await apiClient.getGames();
	return {
		user,
		games,
	};
}

export default function Dashboard() {
	const { user, games } = useLoaderData<typeof loader>()!;

	return (
		<div className="p-6 bg-gray-100 min-h-screen flex flex-col items-center gap-4">
			{user.icon_path && (
				<UserIcon
					iconPath={user.icon_path}
					displayName={user.display_name}
					className="w-24 h-24"
				/>
			)}
			<h1 className="text-3xl font-bold text-gray-800">{user.display_name}</h1>
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
									</div>
								</li>
							))}
						</ul>
					)}
				</div>
			</BorderedContainerWithCaption>
			<Form method="post" action="/logout">
				<button
					type="submit"
					className="px-4 py-2 bg-red-500 text-white rounded-sm transition duration-300 hover:bg-red-700 focus:ring-3 focus:ring-red-400 focus:outline-hidden"
				>
					ログアウト
				</button>
			</Form>
			{user.is_admin && (
				<a
					href={
						process.env.NODE_ENV === "development"
							? `http://localhost:8003${BASE_PATH}admin/dashboard`
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
