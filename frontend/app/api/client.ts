import createClient from "openapi-fetch";
import { createContext } from "react";
import { API_BASE_PATH } from "../config";
import type { paths } from "./schema";

const apiOrigin =
	import.meta.env.VITE_API_BASE_URL ??
	(import.meta.env.DEV ? "http://localhost:8004" : "");

const client = createClient<paths>({
	baseUrl: `${apiOrigin}${API_BASE_PATH}`,
	credentials: "include",
});

export async function apiLogin(username: string, password: string) {
	const { data, error } = await client.POST("/login", {
		body: {
			username,
			password,
		},
	});
	if (error) throw new Error(error.message);
	return data;
}

export async function apiLogout() {
	const { error } = await client.POST("/logout");
	if (error) throw new Error(error.message);
}

export async function apiGetMe() {
	const { data, error } = await client.GET("/me");
	if (error) return null;
	return data;
}

class AuthenticatedApiClient {
	async getGames() {
		const { data, error } = await client.GET("/games");
		if (error) throw new Error(error.message);
		return data;
	}

	async getGame(gameId: number) {
		const { data, error } = await client.GET("/games/{game_id}", {
			params: {
				path: { game_id: gameId },
			},
		});
		if (error) throw new Error(error.message);
		return data;
	}

	async getGamePlayLatestState(gameId: number) {
		const { data, error } = await client.GET(
			"/games/{game_id}/play/latest_state",
			{
				params: {
					path: { game_id: gameId },
				},
			},
		);
		if (error) throw new Error(error.message);
		return data;
	}

	async postGamePlayCode(gameId: number, code: string) {
		const { error } = await client.POST("/games/{game_id}/play/code", {
			params: {
				path: { game_id: gameId },
			},
			body: { code },
		});
		if (error) throw new Error(error.message);
	}

	async postGamePlaySubmit(gameId: number, code: string) {
		const { data, error } = await client.POST("/games/{game_id}/play/submit", {
			params: {
				path: { game_id: gameId },
			},
			body: { code },
		});
		if (error) throw new Error(error.message);
		return data;
	}

	async getGamePlaySubmissions(gameId: number) {
		const { data, error } = await client.GET(
			"/games/{game_id}/play/submissions",
			{
				params: {
					path: { game_id: gameId },
				},
			},
		);
		if (error) throw new Error(error.message);
		return data;
	}

	async getGameWatchRanking(gameId: number) {
		const { data, error } = await client.GET("/games/{game_id}/watch/ranking", {
			params: {
				path: { game_id: gameId },
			},
		});
		if (error) throw new Error(error.message);
		return data;
	}

	async getGameWatchLatestStates(gameId: number) {
		const { data, error } = await client.GET(
			"/games/{game_id}/watch/latest_states",
			{
				params: {
					path: { game_id: gameId },
				},
			},
		);
		if (error) throw new Error(error.message);
		return data;
	}

	async getTournament(tournamentId: number) {
		const { data, error } = await client.GET("/tournaments/{tournament_id}", {
			params: {
				path: { tournament_id: tournamentId },
			},
		});
		if (error) throw new Error(error.message);
		return data;
	}
}

const apiClient = new AuthenticatedApiClient();

export function createApiClient() {
	return apiClient;
}

export const ApiClientContext = createContext<AuthenticatedApiClient | null>(
	null,
);
