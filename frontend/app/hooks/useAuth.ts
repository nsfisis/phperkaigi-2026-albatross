import { useCallback, useSyncExternalStore } from "react";
import { apiLogin } from "../api/client";
import {
	type User,
	clearToken,
	getToken,
	getUserFromToken,
	isTokenExpired,
	setToken,
} from "../auth";

// Simple external store to trigger re-renders when auth state changes.
let authVersion = 0;
const listeners = new Set<() => void>();

function subscribe(callback: () => void) {
	listeners.add(callback);
	return () => listeners.delete(callback);
}

function getSnapshot() {
	return authVersion;
}

function notifyAuthChange() {
	authVersion++;
	for (const listener of listeners) {
		listener();
	}
}

export function useAuth(): {
	user: User | null;
	token: string | null;
	isLoggedIn: boolean;
	login: (username: string, password: string) => Promise<void>;
	logout: () => void;
} {
	useSyncExternalStore(subscribe, getSnapshot);

	const token = getToken();
	const isExpired = isTokenExpired();
	const user = isExpired ? null : getUserFromToken();
	const isLoggedIn = user !== null && !isExpired;

	const login = useCallback(async (username: string, password: string) => {
		const { token } = await apiLogin(username, password);
		setToken(token);
		notifyAuthChange();
	}, []);

	const logout = useCallback(() => {
		clearToken();
		notifyAuthChange();
	}, []);

	return { user, token: isLoggedIn ? token : null, isLoggedIn, login, logout };
}
