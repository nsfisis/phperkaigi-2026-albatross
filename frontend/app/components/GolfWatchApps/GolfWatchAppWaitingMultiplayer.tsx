type Props = {
  gameDisplayName: string;
};

export default function GolfWatchAppWaitingMultiplayer({
  gameDisplayName,
}: Props) {
  return (
    <div className="min-h-screen bg-gray-100 flex flex-col font-bold text-center">
      <div className="text-white bg-brand-600 p-10">
        <div className="text-4xl">{gameDisplayName}</div>
      </div>
    </div>
  );
}
