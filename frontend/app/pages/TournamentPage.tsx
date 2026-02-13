import { useEffect, useState } from "react";
import { createApiClient } from "../api/client";
import type { components } from "../api/schema";
import BorderedContainer from "../components/BorderedContainer";
import UserIcon from "../components/UserIcon";
import { APP_NAME } from "../config";
import { usePageTitle } from "../hooks/usePageTitle";

type TournamentMatch = components["schemas"]["TournamentMatch"];
type User = components["schemas"]["User"];

function Player({ player, rank }: { player: User | null; rank: number }) {
	return (
		<BorderedContainer>
			<div className="flex flex-col items-center gap-2">
				<span className="text-gray-800 text-md">予選 {rank} 位</span>
				<span className="font-medium text-lg">{player?.display_name}</span>
				{player?.icon_path && (
					<UserIcon
						iconPath={player.icon_path}
						displayName={player.display_name}
						className="w-16 h-16 my-auto"
					/>
				)}
			</div>
		</BorderedContainer>
	);
}

function BranchVL({ className = "" }: { className?: string }) {
	return (
		<div className="grid grid-cols-2">
			<div></div>
			<div className={`border-l-4 ${className}`}></div>
		</div>
	);
}

function BranchVR({ className = "" }: { className?: string }) {
	return (
		<div className="grid grid-cols-2">
			<div className={`border-r-4 ${className}`}></div>
			<div></div>
		</div>
	);
}

function BranchVL2({
	score,
	className = "",
}: { score: number | null; className?: string }) {
	return (
		<div className="grid grid-cols-3">
			<div className={`border-r-4 ${className}`}></div>
			<div className={`border-t-4 p-2 font-bold text-xl ${className}`}>
				{score}
			</div>
			<div className={`border-t-4 ${className}`}></div>
		</div>
	);
}

function BranchVR2({
	score,
	className = "",
}: { score: number | null; className?: string }) {
	return (
		<div className="grid grid-cols-3">
			<div className={`border-t-4 ${className}`}></div>
			<div className={`border-t-4 p-2 font-bold text-xl ${className}`}>
				{score}
			</div>
			<div className={`border-l-4 ${className}`}></div>
		</div>
	);
}

function BranchV3({ className = "" }: { className?: string }) {
	return <div className={`border-r-4 ${className}`}></div>;
}

function BranchH({
	score,
	className1,
	className2,
	className3,
}: {
	score?: number | null;
	className1: string;
	className2: string;
	className3: string;
}) {
	return (
		<div className="grid grid-cols-3">
			<div className={`border-t-4 ${className1}`}></div>
			<div className={`border-t-4 ${className2}`}></div>
			<div className={`border-t-4 p-2 font-bold text-xl ${className3}`}>
				{score}
			</div>
		</div>
	);
}

function BranchH2({
	score,
	className1,
	className2,
	className3,
}: {
	score?: number | null;
	className1: string;
	className2: string;
	className3: string;
}) {
	return (
		<div className="grid grid-cols-3">
			<div
				className={`border-t-4 p-2 font-bold text-xl text-right ${className1}`}
			>
				{score}
			</div>
			<div className={`border-t-4 ${className2}`}></div>
			<div className={`border-t-4 ${className3}`}></div>
		</div>
	);
}

function BranchL({
	score,
	className = "",
}: { score: number | null; className?: string }) {
	return (
		<div className="grid grid-cols-2">
			<div></div>
			<div
				className={`border-l-4 border-t-4 p-2 font-bold text-xl ${className}`}
			>
				{score}
			</div>
		</div>
	);
}

function BranchR({
	score,
	className = "",
}: { score: number | null; className?: string }) {
	return (
		<div className="grid grid-cols-2">
			<div
				className={`border-r-4 border-t-4 p-2 font-bold text-xl text-right ${className}`}
			>
				{score}
			</div>
			<div></div>
		</div>
	);
}

function BranchL2({ className = "" }: { className?: string }) {
	return (
		<div className="grid grid-cols-2">
			<div className={`border-l-4 ${className}`}></div>
			<div></div>
		</div>
	);
}

function BranchR2({ className = "" }: { className?: string }) {
	return (
		<div className="grid grid-cols-2">
			<div></div>
			<div className={`border-r-4 ${className}`}></div>
		</div>
	);
}

function getPlayer(match: TournamentMatch, playerID: number): User | null {
	if (match.player1?.user_id === playerID) return match.player1;
	if (match.player2?.user_id === playerID) return match.player2;
	return null;
}

function getScore(match: TournamentMatch, playerIDs: number[]): number | null {
	if (match.player1 && playerIDs.includes(match.player1.user_id))
		return match.player1_score ?? null;
	if (match.player2 && playerIDs.includes(match.player2.user_id))
		return match.player2_score ?? null;
	return null;
}

function getBorderColor(match: TournamentMatch, playerIDs: number[]): string {
	if (!match.winner) {
		return "border-black";
	}
	if (playerIDs.includes(match.winner)) {
		return "border-pink-700";
	}
	return "border-gray-400";
}

export default function TournamentPage() {
	usePageTitle(`Tournament | ${APP_NAME}`);

	const [tournament, setTournament] = useState<{
		matches: TournamentMatch[];
	} | null>(null);
	const [playerIDs, setPlayerIDs] = useState<number[]>([]);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState<string | null>(null);

	useEffect(() => {
		const params = new URLSearchParams(window.location.search);
		const game1 = Number(params.get("game1"));
		const game2 = Number(params.get("game2"));
		const game3 = Number(params.get("game3"));
		const game4 = Number(params.get("game4"));
		const game5 = Number(params.get("game5"));

		if (!game1 || !game2 || !game3 || !game4 || !game5) {
			setError("Missing or invalid game parameters");
			setLoading(false);
			return;
		}

		const pIDs = [
			Number(params.get("player1")),
			Number(params.get("player2")),
			Number(params.get("player3")),
			Number(params.get("player4")),
			Number(params.get("player5")),
			Number(params.get("player6")),
		];

		if (pIDs.some((id) => !id)) {
			setError("Missing or invalid player parameters");
			setLoading(false);
			return;
		}

		setPlayerIDs(pIDs);

		const apiClient = createApiClient();
		apiClient
			.getTournament(game1, game2, game3, game4, game5)
			.then(({ tournament }) => setTournament(tournament))
			.catch(() => setError("Failed to load tournament"))
			.finally(() => setLoading(false));
	}, []);

	if (loading) {
		return (
			<div className="min-h-screen bg-gray-100 flex items-center justify-center">
				<p className="text-gray-500">Loading...</p>
			</div>
		);
	}

	if (error || !tournament) {
		return (
			<div className="min-h-screen bg-gray-100 flex items-center justify-center">
				<p className="text-red-500">{error || "Failed to load tournament"}</p>
			</div>
		);
	}

	const match1 = tournament.matches[0]!;
	const match2 = tournament.matches[1]!;
	const match3 = tournament.matches[2]!;
	const match4 = tournament.matches[3]!;
	const match5 = tournament.matches[4]!;

	const playerID1 = playerIDs[0]!;
	const playerID2 = playerIDs[1]!;
	const playerID3 = playerIDs[2]!;
	const playerID4 = playerIDs[3]!;
	const playerID5 = playerIDs[4]!;
	const playerID6 = playerIDs[5]!;

	const player5 = getPlayer(match1, playerID5);
	const player4 = getPlayer(match1, playerID4);
	const player3 = getPlayer(match2, playerID3);
	const player6 = getPlayer(match2, playerID6);
	const player1 = getPlayer(match3, playerID1);
	const player2 = getPlayer(match4, playerID2);

	return (
		<div className="p-6 bg-gray-100 min-h-screen">
			<div className="max-w-5xl mx-auto">
				<h1 className="text-3xl font-bold text-transparent bg-clip-text bg-iosdc-japan text-center mb-8">
					iOSDC Japan 2025 Swift Code Battle
				</h1>

				<div className="grid grid-rows-5">
					<div className="grid grid-cols-6">
						<div></div>
						<div></div>
						<BranchV3
							className={getBorderColor(match5, [
								playerID1,
								playerID5,
								playerID4,
								playerID3,
								playerID6,
								playerID2,
							])}
						/>
						<div></div>
						<div></div>
						<div></div>
					</div>
					<div className="grid grid-cols-6">
						<div></div>
						<BranchVL2
							score={getScore(match5, [playerID1, playerID5, playerID4])}
							className={getBorderColor(match5, [
								playerID1,
								playerID5,
								playerID4,
							])}
						/>
						<BranchH
							className1={getBorderColor(match5, [
								playerID1,
								playerID5,
								playerID4,
							])}
							className2={getBorderColor(match5, [
								playerID1,
								playerID5,
								playerID4,
							])}
							className3={getBorderColor(match5, [
								playerID1,
								playerID5,
								playerID4,
							])}
						/>
						<BranchH
							className1={getBorderColor(match5, [
								playerID3,
								playerID6,
								playerID2,
							])}
							className2={getBorderColor(match5, [
								playerID3,
								playerID6,
								playerID2,
							])}
							className3={getBorderColor(match5, [
								playerID3,
								playerID6,
								playerID2,
							])}
						/>
						<BranchVR2
							score={getScore(match5, [playerID3, playerID6, playerID2])}
							className={getBorderColor(match5, [
								playerID3,
								playerID6,
								playerID2,
							])}
						/>
						<div></div>
					</div>
					<div className="grid grid-cols-6">
						<BranchL
							score={getScore(match3, [playerID1])}
							className={getBorderColor(match3, [playerID1])}
						/>
						<BranchH
							score={getScore(match3, [playerID5, playerID4])}
							className1={getBorderColor(match3, [playerID1])}
							className2={getBorderColor(match3, [playerID5, playerID4])}
							className3={getBorderColor(match3, [playerID5, playerID4])}
						/>
						<BranchL2
							className={getBorderColor(match3, [playerID5, playerID4])}
						/>
						<BranchR2
							className={getBorderColor(match4, [playerID3, playerID6])}
						/>
						<BranchH2
							score={getScore(match4, [playerID3, playerID6])}
							className1={getBorderColor(match4, [playerID3, playerID6])}
							className2={getBorderColor(match4, [playerID3, playerID6])}
							className3={getBorderColor(match4, [playerID2])}
						/>
						<BranchR
							score={getScore(match4, [playerID2])}
							className={getBorderColor(match4, [playerID2])}
						/>
					</div>
					<div className="grid grid-cols-6">
						<BranchVL className={getBorderColor(match3, [playerID1])} />
						<BranchL
							score={getScore(match1, [playerID5])}
							className={getBorderColor(match1, [playerID5])}
						/>
						<BranchR
							score={getScore(match1, [playerID4])}
							className={getBorderColor(match1, [playerID4])}
						/>
						<BranchL
							score={getScore(match2, [playerID3])}
							className={getBorderColor(match2, [playerID3])}
						/>
						<BranchR
							score={getScore(match2, [playerID6])}
							className={getBorderColor(match2, [playerID6])}
						/>
						<BranchVR className={getBorderColor(match4, [playerID2])} />
					</div>
					<div className="grid grid-cols-6 gap-6">
						<Player player={player1} rank={1} />
						<Player player={player5} rank={5} />
						<Player player={player4} rank={4} />
						<Player player={player3} rank={3} />
						<Player player={player6} rank={6} />
						<Player player={player2} rank={2} />
					</div>
				</div>
			</div>
		</div>
	);
}
