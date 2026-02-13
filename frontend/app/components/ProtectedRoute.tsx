import { Redirect } from "wouter";
import { useAuth } from "../hooks/useAuth";

export default function ProtectedRoute({
	children,
}: {
	children: React.ReactNode;
}) {
	const { isLoggedIn, isLoading } = useAuth();

	if (isLoading) {
		return null;
	}

	if (!isLoggedIn) {
		return <Redirect to="/login" />;
	}

	return <>{children}</>;
}
