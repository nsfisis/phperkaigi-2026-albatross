import { Redirect } from "wouter";
import { useAuth } from "../hooks/useAuth";

export default function PublicOnlyRoute({
	children,
}: {
	children: React.ReactNode;
}) {
	const { isLoggedIn, isLoading } = useAuth();

	if (isLoading) {
		return null;
	}

	if (isLoggedIn) {
		return <Redirect to="/dashboard" />;
	}

	return <>{children}</>;
}
