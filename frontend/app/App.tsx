import { Route, Router, Switch } from "wouter";
import ProtectedRoute from "./components/ProtectedRoute";
import PublicOnlyRoute from "./components/PublicOnlyRoute";
import { BASE_PATH } from "./config";
import DashboardPage from "./pages/DashboardPage";
import GolfPlayPage from "./pages/GolfPlayPage";
import GolfProblemPreviewPage from "./pages/GolfProblemPreviewPage";
import GolfWatchPage from "./pages/GolfWatchPage";
import IndexPage from "./pages/IndexPage";
import LoginPage from "./pages/LoginPage";
import SubmissionsPage from "./pages/SubmissionsPage";
import TournamentPage from "./pages/TournamentPage";

export default function App() {
	return (
		<Router base={BASE_PATH.replace(/\/$/, "")}>
			<Switch>
				<Route path="/">
					<PublicOnlyRoute>
						<IndexPage />
					</PublicOnlyRoute>
				</Route>
				<Route path="/login">
					<PublicOnlyRoute>
						<LoginPage />
					</PublicOnlyRoute>
				</Route>
				<Route path="/dashboard">
					<DashboardPage />
				</Route>
				<Route path="/golf/:gameId/preview">
					{(params) => (
						<ProtectedRoute>
							<GolfProblemPreviewPage gameId={params.gameId} />
						</ProtectedRoute>
					)}
				</Route>
				<Route path="/golf/:gameId/play">
					{(params) => (
						<ProtectedRoute>
							<GolfPlayPage gameId={params.gameId} />
						</ProtectedRoute>
					)}
				</Route>
				<Route path="/golf/:gameId/submissions">
					{(params) => (
						<ProtectedRoute>
							<SubmissionsPage gameId={params.gameId} />
						</ProtectedRoute>
					)}
				</Route>
				<Route path="/golf/:gameId/watch">
					{(params) => <GolfWatchPage gameId={params.gameId} />}
				</Route>
				<Route path="/tournament/:tournamentId">
					{(params) => <TournamentPage tournamentId={params.tournamentId} />}
				</Route>
				<Route>
					<div className="min-h-screen bg-gray-100 flex items-center justify-center">
						<p className="text-gray-500 text-xl">404 - Page not found</p>
					</div>
				</Route>
			</Switch>
		</Router>
	);
}
