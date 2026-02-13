import { type JwtPayload, jwtDecode } from "jwt-decode";
import type { components } from "./api/schema";

export type User = components["schemas"]["User"];

const COOKIE_NAME = "albatross_token";

export function getToken(): string | null {
	const match = document.cookie
		.split("; ")
		.find((row) => row.startsWith(`${COOKIE_NAME}=`));
	if (!match) return null;
	return match.split("=").slice(1).join("=");
}

export function setToken(token: string): void {
	document.cookie = `${COOKIE_NAME}=${token}; path=/; SameSite=Lax`;
}

export function clearToken(): void {
	document.cookie = `${COOKIE_NAME}=; path=/; SameSite=Lax; max-age=0`;
}

export function getUserFromToken(): User | null {
	const token = getToken();
	if (!token) return null;
	try {
		return jwtDecode<User & JwtPayload>(token);
	} catch {
		return null;
	}
}

export function isTokenExpired(): boolean {
	const token = getToken();
	if (!token) return true;
	try {
		const decoded = jwtDecode<JwtPayload>(token);
		if (decoded.exp == null) return false;
		// If the token will expire in less than an hour, treat it as expired.
		return new Date((decoded.exp - 3600) * 1000) < new Date();
	} catch {
		return true;
	}
}
