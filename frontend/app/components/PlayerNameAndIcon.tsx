import { PlayerProfile } from "../types/PlayerProfile";
import UserIcon from "./UserIcon";

type Props = {
  profile: PlayerProfile;
};

export default function PlayerNameAndIcon({ profile }: Props) {
  return (
    <div className="flex flex-col gap-6 my-auto items-center">
      <div className="text-6xl">{profile.displayName}</div>
      {profile.iconPath && (
        <UserIcon
          iconPath={profile.iconPath}
          displayName={profile.displayName}
          className="w-48 h-48"
        />
      )}
    </div>
  );
}
