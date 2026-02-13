import { useCallback, useEffect, useState } from "react";
import { apiGetMe, apiLogin, apiLogout } from "../api/client";
import type { User } from "../auth";

export function useAuth(): {
	user: User | null;
	isLoggedIn: boolean;
	isLoading: boolean;
	login: (username: string, password: string) => Promise<void>;
	logout: () => Promise<void>;
} {
	const [user, setUser] = useState<User | null>(null);
	const [isLoading, setIsLoading] = useState(true);

	useEffect(() => {
		apiGetMe()
			.then((data) => setUser(data?.user ?? null))
			.catch(() => setUser(null))
			.finally(() => setIsLoading(false));
	}, []);

	const login = useCallback(async (username: string, password: string) => {
		const { user } = await apiLogin(username, password);
		setUser(user);
	}, []);

	const logout = useCallback(async () => {
		await apiLogout();
		setUser(null);
	}, []);

	return { user, isLoggedIn: user !== null, isLoading, login, logout };
}
