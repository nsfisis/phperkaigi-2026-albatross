import type { SupportedLanguage } from "../../types/SupportedLanguage";
import TitledColumn from "../TitledColumn";
import ProblemColumnContent from "./ProblemColumnContent";

type Props = {
	title: string;
	description: string;
	language: SupportedLanguage;
	sampleCode: string;
};

export default function ProblemColumn({
	title,
	description,
	language,
	sampleCode,
}: Props) {
	return (
		<TitledColumn title={title}>
			<ProblemColumnContent
				description={description}
				sampleCode={sampleCode}
				language={language}
			/>
		</TitledColumn>
	);
}
