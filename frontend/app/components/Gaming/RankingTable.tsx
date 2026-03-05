import { useAtomValue } from "jotai";
import { rankingAtom } from "../../states/watch";
import type { SupportedLanguage } from "../../types/SupportedLanguage";
import CodePopover from "./CodePopover";
import DataTable, { DataTableCell, formatUnixTimestamp } from "./DataTable";

type Props = {
	problemLanguage: SupportedLanguage;
};

export default function RankingTable({ problemLanguage }: Props) {
	const ranking = useAtomValue(rankingAtom);
	const showCode = ranking.some((entry) => entry.code != null);

	return (
		<DataTable
			headers={[
				"順位",
				"プレイヤー",
				"スコア",
				"提出時刻",
				...(showCode ? ["コード"] : []),
			]}
		>
			{ranking.map((entry, index) => (
				<tr key={entry.player.user_id}>
					<DataTableCell>{index + 1}</DataTableCell>
					<DataTableCell>
						{entry.player.display_name}
						{entry.player.label && ` (${entry.player.label})`}
					</DataTableCell>
					<DataTableCell>{entry.score}</DataTableCell>
					<DataTableCell>
						{formatUnixTimestamp(entry.submitted_at)}
					</DataTableCell>
					{showCode && (
						<DataTableCell>
							{entry.code && (
								<CodePopover code={entry.code} language={problemLanguage} />
							)}
						</DataTableCell>
					)}
				</tr>
			))}
		</DataTable>
	);
}
