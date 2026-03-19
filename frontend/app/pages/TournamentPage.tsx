import { useEffect, useState } from "react";
import { createApiClient } from "../api/client";
import type { components } from "../api/schema";
import BorderedContainer from "../components/BorderedContainer";
import UserIcon from "../components/UserIcon";
import { APP_NAME } from "../config";
import { usePageTitle } from "../hooks/usePageTitle";

type Tournament = components["schemas"]["Tournament"];
type TournamentMatch = components["schemas"]["TournamentMatch"];
type TournamentEntry = components["schemas"]["TournamentEntry"];

function getBorderColor(match: TournamentMatch, userID?: number): string {
  if (!match.winner_user_id) {
    return "border-black";
  }
  if (userID !== undefined && match.winner_user_id === userID) {
    return "border-brand-600";
  }
  return "border-gray-400";
}

function getBgColor(match: TournamentMatch, userID?: number): string {
  if (!match.winner_user_id) {
    return "bg-black";
  }
  if (userID !== undefined && match.winner_user_id === userID) {
    return "bg-brand-600";
  }
  return "bg-gray-400";
}

function getWinnerBgColor(match: TournamentMatch | undefined): string {
  if (!match?.winner_user_id) {
    return "bg-black";
  }
  return "bg-brand-600";
}

function PlayerCard({ entry }: { entry: TournamentEntry | undefined }) {
  if (!entry) {
    return (
      <div className="flex flex-col items-center gap-1 p-2 opacity-30">
        <span className="text-gray-400 text-sm">BYE</span>
      </div>
    );
  }
  return (
    <BorderedContainer>
      <div className="flex flex-col items-center gap-1">
        <span className="font-medium text-sm truncate max-w-full">
          {entry.user.display_name}
        </span>
        {entry.user.icon_path && (
          <UserIcon
            iconPath={entry.user.icon_path}
            displayName={entry.user.display_name}
            className="w-12 h-12"
          />
        )}
      </div>
    </BorderedContainer>
  );
}

function Connector({
  position,
  colSpan,
  match,
  isTopRound,
  leftChildMatch,
  rightChildMatch,
}: {
  position: number;
  colSpan: number;
  match: TournamentMatch | undefined;
  isTopRound: boolean;
  leftChildMatch?: TournamentMatch;
  rightChildMatch?: TournamentMatch;
}) {
  const leftHalf = colSpan / 2;
  const rightHalf = colSpan - leftHalf;

  const leftBorderColor = match
    ? getBorderColor(match, match.player1?.user_id)
    : "border-black";
  const rightBorderColor = match
    ? getBorderColor(match, match.player2?.user_id)
    : "border-black";
  // Leg colors: use child match winner color so vertical lines don't change color midway
  const leftLegColor = leftChildMatch
    ? getWinnerBgColor(leftChildMatch)
    : match
      ? getBgColor(match, match.player1?.user_id)
      : "bg-black";
  const rightLegColor = rightChildMatch
    ? getWinnerBgColor(rightChildMatch)
    : match
      ? getBgColor(match, match.player2?.user_id)
      : "bg-black";
  const winnerBg = getWinnerBgColor(match);

  return (
    <div
      style={{
        gridColumn: `${position * colSpan + 1} / span ${colSpan}`,
      }}
    >
      {/* Center stem going UP to parent round */}
      <div className={`flex justify-center ${isTopRound ? "h-24" : "h-12"}`}>
        <div className={`w-0.5 ${winnerBg}`} />
      </div>

      {/* Horizontal line */}
      <div className="flex justify-center">
        <div className={`w-1/4 border-t-2 ${leftBorderColor}`} />
        <div className={`w-1/4 border-t-2 ${rightBorderColor}`} />
      </div>

      {/* Two legs going DOWN to child elements */}
      <div
        className="grid h-12"
        style={{
          gridTemplateColumns: `repeat(${colSpan}, 1fr)`,
        }}
      >
        <div
          className="flex justify-center"
          style={{ gridColumn: `1 / span ${leftHalf}` }}
        >
          <div className={`w-0.5 ${leftLegColor}`} />
        </div>
        <div
          className="flex justify-center"
          style={{ gridColumn: `${leftHalf + 1} / span ${rightHalf}` }}
        >
          <div className={`w-0.5 ${rightLegColor}`} />
        </div>
      </div>
    </div>
  );
}

function TournamentBracket({ tournament }: { tournament: Tournament }) {
  const { bracket_size, num_rounds, entries, matches } = tournament;

  const matchByKey = new Map<string, TournamentMatch>();
  for (const m of matches) {
    matchByKey.set(`${m.round}-${m.position}`, m);
  }

  const entryBySeed = new Map<number, TournamentEntry>();
  for (const e of entries) {
    entryBySeed.set(e.seed, e);
  }

  const bracketSeeds = standardBracketSeeds(bracket_size);

  // Build rows top-to-bottom: final → ... → round 0 → players
  const rows: React.ReactNode[] = [];

  // Rounds from top (final) to bottom (round 0)
  for (let round = num_rounds - 1; round >= 0; round--) {
    const numPositions = bracket_size / (1 << (round + 1));
    const colSpan = bracket_size / numPositions;

    // Connectors for this round
    const connectors: React.ReactNode[] = [];
    for (let pos = 0; pos < numPositions; pos++) {
      const match = matchByKey.get(`${round}-${pos}`);
      const leftChildMatch =
        round > 0 ? matchByKey.get(`${round - 1}-${pos * 2}`) : undefined;
      const rightChildMatch =
        round > 0 ? matchByKey.get(`${round - 1}-${pos * 2 + 1}`) : undefined;
      connectors.push(
        <Connector
          key={`conn-${round}-${pos}`}
          position={pos}
          colSpan={colSpan}
          match={match}
          isTopRound={round === num_rounds - 1}
          leftChildMatch={leftChildMatch}
          rightChildMatch={rightChildMatch}
        />,
      );
    }
    rows.push(
      <div
        key={`conn-row-${round}`}
        className="grid"
        style={{
          gridTemplateColumns: `repeat(${bracket_size}, 1fr)`,
        }}
      >
        {connectors}
      </div>,
    );
  }

  // Player stems row (vertical lines from round-0 connectors to player cards)
  const playerStems: React.ReactNode[] = [];
  for (let slot = 0; slot < bracket_size; slot++) {
    const matchPos = Math.floor(slot / 2);
    const match = matchByKey.get(`0-${matchPos}`);
    const isPlayer1 = slot % 2 === 0;
    const stemColor = match
      ? getBgColor(
          match,
          isPlayer1 ? match.player1?.user_id : match.player2?.user_id,
        )
      : "bg-black";
    playerStems.push(
      <div
        key={`stem-${slot}`}
        className="flex justify-center h-12"
        style={{ gridColumn: `${slot + 1} / span 1` }}
      >
        <div className={`w-0.5 ${stemColor}`} />
      </div>,
    );
  }
  rows.push(
    <div
      key="player-stems"
      className="grid"
      style={{ gridTemplateColumns: `repeat(${bracket_size}, 1fr)` }}
    >
      {playerStems}
    </div>,
  );

  // Player cards row (bottom)
  const playerCards: React.ReactNode[] = [];
  for (let slot = 0; slot < bracket_size; slot++) {
    const seed = bracketSeeds[slot]!;
    const entry = entryBySeed.get(seed);
    playerCards.push(
      <div
        key={`player-${slot}`}
        style={{ gridColumn: `${slot + 1} / span 1` }}
      >
        <PlayerCard entry={entry} />
      </div>,
    );
  }
  rows.push(
    <div
      key="players"
      className="grid gap-1"
      style={{ gridTemplateColumns: `repeat(${bracket_size}, 1fr)` }}
    >
      {playerCards}
    </div>,
  );

  return <div className="flex flex-col gap-0">{rows}</div>;
}

// Exported for testing as standardBracketSeedsForTest
export { standardBracketSeeds as standardBracketSeedsForTest };

function standardBracketSeeds(bracketSize: number): number[] {
  const seeds = new Array<number>(bracketSize).fill(0);
  seeds[0] = 1;
  for (let size = 2; size <= bracketSize; size *= 2) {
    const temp = new Array<number>(size).fill(0);
    for (let i = 0; i < size / 2; i++) {
      temp[i * 2] = seeds[i]!;
      temp[i * 2 + 1] = size + 1 - seeds[i]!;
    }
    for (let i = 0; i < size; i++) {
      seeds[i] = temp[i]!;
    }
  }
  return seeds;
}

export default function TournamentPage({
  tournamentId,
}: {
  tournamentId: string;
}) {
  usePageTitle(`Tournament | ${APP_NAME}`);

  const id = Number(tournamentId);
  const isValidId = id > 0;

  const [tournament, setTournament] = useState<Tournament | null>(null);
  const [loading, setLoading] = useState(isValidId);
  const [error, setError] = useState<string | null>(
    isValidId ? null : "Invalid tournament ID",
  );

  useEffect(() => {
    if (!isValidId) {
      return;
    }

    const apiClient = createApiClient();
    apiClient
      .getTournament(id)
      .then(({ tournament }) => setTournament(tournament))
      .catch(() => setError("Failed to load tournament"))
      .finally(() => setLoading(false));
  }, [id, isValidId]);

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

  return (
    <div className="p-6 bg-gray-100 min-h-screen">
      <div className="max-w-6xl mx-auto">
        <h1 className="text-3xl font-bold text-transparent bg-clip-text bg-brand-600 text-center mb-8">
          {tournament.display_name}
        </h1>
        <TournamentBracket tournament={tournament} />
      </div>
    </div>
  );
}
