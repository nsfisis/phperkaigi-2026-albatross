import type { PlayerProfile } from "../../types/PlayerProfile";
import PlayerNameAndIcon from "../PlayerNameAndIcon";

type Props = {
  gameDisplayName: string;
  playerProfile: PlayerProfile;
};

export default function GolfPlayAppWaiting({
  gameDisplayName,
  playerProfile,
}: Props) {
  return (
    <div className="min-h-screen bg-gray-100 flex flex-col font-bold text-center">
      <div className="text-white bg-brand-600 p-10">
        <div className="text-4xl">{gameDisplayName}</div>
      </div>
      <div className="grow grid mx-auto text-black">
        <PlayerNameAndIcon profile={playerProfile} />
      </div>
    </div>
  );
}
