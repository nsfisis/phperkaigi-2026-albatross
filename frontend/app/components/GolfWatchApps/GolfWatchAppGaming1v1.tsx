import { useAtomValue } from "jotai";
import {
	calcCodeSize,
	checkGameResultKind,
	gameStateKindAtom,
	gamingLeftTimeSecondsAtom,
	latestGameStatesAtom,
} from "../../states/watch";
import type { PlayerProfile } from "../../types/PlayerProfile";
import type { SupportedLanguage } from "../../types/SupportedLanguage";
import FoldableBorderedContainerWithCaption from "../FoldableBorderedContainerWithCaption";
import CodeBlock from "../Gaming/CodeBlock";
import LeftTime from "../Gaming/LeftTime";
import ProblemColumnContent from "../Gaming/ProblemColumnContent";
import RankingTable from "../Gaming/RankingTable";
import Score from "../Gaming/Score";
import ScoreBar from "../Gaming/ScoreBar";
import SubmitStatusLabel from "../SubmitStatusLabel";
import ThreeColumnLayout from "../ThreeColumnLayout";
import TitledColumn from "../TitledColumn";
import UserIcon from "../UserIcon";

type Props = {
	gameDisplayName: string;
	playerProfileA: PlayerProfile | null;
	playerProfileB: PlayerProfile | null;
	problemTitle: string;
	problemDescription: string;
	problemLanguage: SupportedLanguage;
	sampleCode: string;
};

export default function GolfWatchAppGaming1v1({
	gameDisplayName,
	playerProfileA,
	playerProfileB,
	problemTitle,
	problemDescription,
	problemLanguage,
	sampleCode,
}: Props) {
	const gameStateKind = useAtomValue(gameStateKindAtom);
	const leftTimeSeconds = useAtomValue(gamingLeftTimeSecondsAtom)!;
	const latestGameStates = useAtomValue(latestGameStatesAtom);

	const stateA =
		playerProfileA && (latestGameStates[`${playerProfileA.id}`] ?? null);
	const codeA = stateA?.code ?? "";
	const scoreA = stateA?.score ?? null;
	const statusA = stateA?.status ?? "none";
	const stateB =
		playerProfileB && (latestGameStates[`${playerProfileB.id}`] ?? null);
	const codeB = stateB?.code ?? "";
	const scoreB = stateB?.score ?? null;
	const statusB = stateB?.status ?? "none";

	const codeSizeA = calcCodeSize(codeA, problemLanguage);
	const codeSizeB = calcCodeSize(codeB, problemLanguage);

	const gameResultKind = checkGameResultKind(gameStateKind, stateA, stateB);

	const topBg = gameResultKind
		? gameResultKind === "winA"
			? "bg-orange-400"
			: gameResultKind === "winB"
				? "bg-purple-400"
				: "bg-sky-600"
		: "bg-sky-600";

	return (
		<div className="min-h-screen bg-gray-100 flex flex-col">
			<div className={`text-white ${topBg} grid grid-cols-3 px-4 py-2`}>
				<div className="font-bold flex gap-4 justify-start md:justify-between items-center my-auto">
					<div className="flex gap-6 items-center">
						{playerProfileA?.iconPath && (
							<UserIcon
								iconPath={playerProfileA.iconPath}
								displayName={playerProfileA.displayName}
								className="w-12 h-12 my-auto"
							/>
						)}
						<div className="hidden md:block text-4xl">
							{playerProfileA?.displayName}
						</div>
					</div>
					<div className="text-2xl md:text-6xl">
						<Score status={statusA} score={scoreA} />
					</div>
				</div>
				<div className="font-bold text-center">
					<div className="text-gray-100">{gameDisplayName}</div>
					{gameResultKind ? (
						<div className="text-3xl">
							{gameResultKind === "winA"
								? `勝者 ${playerProfileA!.displayName}`
								: gameResultKind === "winB"
									? `勝者 ${playerProfileB!.displayName}`
									: "引き分け"}
						</div>
					) : (
						<LeftTime sec={leftTimeSeconds} />
					)}
				</div>
				<div className="font-bold flex gap-4 justify-end md:justify-between items-center my-auto">
					<div className="text-2xl md:text-6xl">
						<Score status={statusB} score={scoreB} />
					</div>
					<div className="flex gap-6 items-center text-end">
						<div className="hidden md:block text-4xl">
							{playerProfileB?.displayName}
						</div>
						{playerProfileB?.iconPath && (
							<UserIcon
								iconPath={playerProfileB.iconPath}
								displayName={playerProfileB.displayName}
								className="w-12 h-12 my-auto"
							/>
						)}
					</div>
				</div>
			</div>
			<ScoreBar
				scoreA={scoreA}
				scoreB={scoreB}
				bgA="bg-orange-400"
				bgB="bg-purple-400"
			/>
			<ThreeColumnLayout>
				<TitledColumn
					title={<SubmitStatusLabel status={statusA} />}
					className="order-2 md:order-1"
				>
					<FoldableBorderedContainerWithCaption
						caption={`コードサイズ: ${codeSizeA}`}
					>
						<CodeBlock code={codeA} language={problemLanguage} />
					</FoldableBorderedContainerWithCaption>
				</TitledColumn>
				<TitledColumn title={problemTitle} className="order-1 md:order-2">
					<ProblemColumnContent
						description={problemDescription}
						language={problemLanguage}
						sampleCode={sampleCode}
					/>
					<RankingTable problemLanguage={problemLanguage} />
				</TitledColumn>
				<TitledColumn
					title={<SubmitStatusLabel status={statusB} />}
					className="order-3"
				>
					<FoldableBorderedContainerWithCaption
						caption={`コードサイズ: ${codeSizeB}`}
					>
						<CodeBlock code={codeB} language={problemLanguage} />
					</FoldableBorderedContainerWithCaption>
				</TitledColumn>
			</ThreeColumnLayout>
		</div>
	);
}
