import { useEffect, useState } from "react";
import { createApiClient } from "../api/client";
import type { components } from "../api/schema";
import BorderedContainerWithCaption from "../components/BorderedContainerWithCaption";
import NavigateLink from "../components/NavigateLink";
import SubmitStatusLabel from "../components/SubmitStatusLabel";
import { APP_NAME } from "../config";
import { usePageTitle } from "../hooks/usePageTitle";

type Submission = components["schemas"]["Submission"];

export default function SubmissionsPage({ gameId }: { gameId: string }) {
	usePageTitle(`Submissions | ${APP_NAME}`);

	const [submissions, setSubmissions] = useState<Submission[]>([]);
	const [loading, setLoading] = useState(true);
	const [expandedId, setExpandedId] = useState<number | null>(null);

	const numericGameId = Number(gameId);

	useEffect(() => {
		const apiClient = createApiClient();
		apiClient
			.getGamePlaySubmissions(numericGameId)
			.then(({ submissions }) => setSubmissions(submissions))
			.catch(() => {})
			.finally(() => setLoading(false));
	}, [numericGameId]);

	if (loading) {
		return (
			<div className="min-h-screen bg-gray-100 flex items-center justify-center">
				<p className="text-gray-500">Loading...</p>
			</div>
		);
	}

	return (
		<div className="p-6 bg-gray-100 min-h-screen flex flex-col items-center gap-4">
			<BorderedContainerWithCaption caption="提出履歴">
				<div className="px-4">
					{submissions.length === 0 ? (
						<p>提出履歴はありません</p>
					) : (
						<ul className="divide-y divide-gray-300">
							{submissions.map((s) => (
								<li key={s.submission_id} className="py-3">
									<div className="flex justify-between items-center gap-4">
										<div className="flex items-center gap-3">
											<StatusBadge status={s.status} />
											<span className="font-mono text-lg font-bold">
												{s.code_size}
												<span className="text-sm font-normal text-gray-500 ml-1">
													bytes
												</span>
											</span>
										</div>
										<div className="flex items-center gap-3">
											<span className="text-sm text-gray-500">
												{formatDate(s.created_at)}
											</span>
											<button
												type="button"
												onClick={() =>
													setExpandedId(
														expandedId === s.submission_id
															? null
															: s.submission_id,
													)
												}
												className="text-sm text-sky-600 hover:text-sky-800 underline"
											>
												{expandedId === s.submission_id
													? "コードを隠す"
													: "コードを見る"}
											</button>
										</div>
									</div>
									{expandedId === s.submission_id && (
										<pre className="mt-2 p-3 bg-gray-800 text-gray-100 rounded text-sm overflow-x-auto">
											{s.code}
										</pre>
									)}
								</li>
							))}
						</ul>
					)}
				</div>
			</BorderedContainerWithCaption>
			<NavigateLink to={`/golf/${gameId}/play`}>対戦に戻る</NavigateLink>
			<NavigateLink to="/dashboard">ダッシュボードに戻る</NavigateLink>
		</div>
	);
}

function StatusBadge({
	status,
}: {
	status: components["schemas"]["ExecutionStatus"];
}) {
	const colorClass =
		status === "success"
			? "bg-green-100 text-green-800"
			: status === "running"
				? "bg-yellow-100 text-yellow-800"
				: status === "none"
					? "bg-gray-100 text-gray-800"
					: "bg-red-100 text-red-800";

	return (
		<span className={`px-2 py-1 rounded text-sm font-medium ${colorClass}`}>
			<SubmitStatusLabel status={status} />
		</span>
	);
}

function formatDate(unixTimestamp: number): string {
	const date = new Date(unixTimestamp * 1000);
	return date.toLocaleString("ja-JP", {
		month: "2-digit",
		day: "2-digit",
		hour: "2-digit",
		minute: "2-digit",
		second: "2-digit",
	});
}
