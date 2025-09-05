import { useAtomValue } from "jotai";
import { gamingLeftTimeSecondsAtom } from "../../states/watch";
import type { SupportedLanguage } from "../../types/SupportedLanguage";
import LeftTime from "../Gaming/LeftTime";
import ProblemColumn from "../Gaming/ProblemColumn";
import RankingTable from "../Gaming/RankingTable";
import TitledColumn from "../TitledColumn";
import TwoColumnLayout from "../TwoColumnLayout";

type Props = {
	gameDisplayName: string;
	problemTitle: string;
	problemDescription: string;
	problemLanguage: SupportedLanguage;
	sampleCode: string;
};

export default function GolfWatchAppGamingMultiplayer({
	gameDisplayName,
	problemTitle,
	problemDescription,
	problemLanguage,
	sampleCode,
}: Props) {
	const leftTimeSeconds = useAtomValue(gamingLeftTimeSecondsAtom)!;

	return (
		<div className="min-h-screen bg-gray-100 flex flex-col">
			<div className="text-white bg-sky-600 grid grid-cols-3 px-4 py-2">
				<div className="font-bold flex justify-between my-auto"></div>
				<div className="font-bold text-center">
					<div className="text-gray-100">{gameDisplayName}</div>
					<LeftTime sec={leftTimeSeconds} />
				</div>
				<div className="font-bold flex justify-between my-auto"></div>
			</div>
			<TwoColumnLayout>
				<ProblemColumn
					title={problemTitle}
					description={problemDescription}
					language={problemLanguage}
					sampleCode={sampleCode}
				/>
				<TitledColumn title="順位表">
					<RankingTable problemLanguage={problemLanguage} />
				</TitledColumn>
			</TwoColumnLayout>
		</div>
	);
}
