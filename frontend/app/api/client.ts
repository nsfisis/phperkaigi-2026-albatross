import createClient from "openapi-fetch";
import { createContext } from "react";
import { API_BASE_PATH } from "../config";
import type { paths } from "./schema";

const client = createClient<paths>({
	baseUrl:
		process.env.NODE_ENV === "development"
			? `http://localhost:8004${API_BASE_PATH}`
			: `https://t.nil.ninja${API_BASE_PATH}`,
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

class AuthenticatedApiClient {
	constructor(public readonly token: string) {}

	async getGames() {
		const { data, error } = await client.GET("/games", {
			params: {
				header: this._getAuthorizationHeader(),
			},
		});
		if (error) throw new Error(error.message);
		return data;
	}

	async getGame(gameId: number) {
		const { data, error } = await client.GET("/games/{game_id}", {
			params: {
				header: this._getAuthorizationHeader(),
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
					header: this._getAuthorizationHeader(),
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
				header: this._getAuthorizationHeader(),
				path: { game_id: gameId },
			},
			body: { code },
		});
		if (error) throw new Error(error.message);
	}

	async postGamePlaySubmit(gameId: number, code: string) {
		const { data, error } = await client.POST("/games/{game_id}/play/submit", {
			params: {
				header: this._getAuthorizationHeader(),
				path: { game_id: gameId },
			},
			body: { code },
		});
		if (error) throw new Error(error.message);
		return data;
	}

	async getGameWatchRanking(gameId: number) {
		const { data, error } = await client.GET("/games/{game_id}/watch/ranking", {
			params: {
				header: this._getAuthorizationHeader(),
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
					header: this._getAuthorizationHeader(),
					path: { game_id: gameId },
				},
			},
		);
		if (error) throw new Error(error.message);
		return data;
	}

	_getAuthorizationHeader() {
		return { Authorization: `Bearer ${this.token}` };
	}
}

export function createApiClient(token: string) {
	return new AuthenticatedApiClient(token);
}

export const ApiClientContext = createContext<AuthenticatedApiClient | null>(
	null,
);
