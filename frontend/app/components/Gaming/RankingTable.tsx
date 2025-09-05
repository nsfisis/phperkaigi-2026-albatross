import { useAtomValue } from "jotai";
import React from "react";
import { rankingAtom } from "../../states/watch";
import type { SupportedLanguage } from "../../types/SupportedLanguage";
import CodePopover from "./CodePopover";

function TableHeaderCell({ children }: { children: React.ReactNode }) {
	return (
		<th scope="col" className="px-6 py-3 text-left font-medium text-gray-800">
			{children}
		</th>
	);
}

function TableBodyCell({ children }: { children: React.ReactNode }) {
	return (
		<td className="px-6 py-4 whitespace-nowrap text-gray-900">{children}</td>
	);
}

function formatUnixTimestamp(timestamp: number) {
	const date = new Date(timestamp * 1000);

	const year = date.getFullYear();
	const month = (date.getMonth() + 1).toString().padStart(2, "0");
	const day = date.getDate().toString().padStart(2, "0");
	const hours = date.getHours().toString().padStart(2, "0");
	const minutes = date.getMinutes().toString().padStart(2, "0");

	return `${year}-${month}-${day} ${hours}:${minutes}`;
}

type Props = {
	problemLanguage: SupportedLanguage;
};

export default function RankingTable({ problemLanguage }: Props) {
	const ranking = useAtomValue(rankingAtom);

	return (
		<div className="overflow-x-auto border-2 border-blue-600 rounded-xl">
			<table className="min-w-full divide-y divide-gray-400 border-collapse">
				<thead className="bg-gray-50">
					<tr>
						<TableHeaderCell>順位</TableHeaderCell>
						<TableHeaderCell>プレイヤー</TableHeaderCell>
						<TableHeaderCell>スコア</TableHeaderCell>
						<TableHeaderCell>提出時刻</TableHeaderCell>
						<TableHeaderCell>コード</TableHeaderCell>
					</tr>
				</thead>
				<tbody className="bg-white divide-y divide-gray-300">
					{ranking.map((entry, index) => (
						<tr key={entry.player.user_id}>
							<TableBodyCell>{index + 1}</TableBodyCell>
							<TableBodyCell>
								{entry.player.display_name}
								{entry.player.label && ` (${entry.player.label})`}
							</TableBodyCell>
							<TableBodyCell>{entry.score}</TableBodyCell>
							<TableBodyCell>
								{formatUnixTimestamp(entry.submitted_at)}
							</TableBodyCell>
							<TableBodyCell>
								{entry.code && (
									<CodePopover code={entry.code} language={problemLanguage} />
								)}
							</TableBodyCell>
						</tr>
					))}
				</tbody>
			</table>
		</div>
	);
}
