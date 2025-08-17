import type { LoaderFunctionArgs, MetaFunction } from "react-router";
import { ensureUserNotLoggedIn } from "../.server/auth";
import BorderedContainer from "../components/BorderedContainer";
import NavigateLink from "../components/NavigateLink";
import { BASE_PATH } from "../config";

export const meta: MetaFunction = () => [
	{ title: "iOSDC Japan 2025 Albatross" },
];

export async function loader({ request }: LoaderFunctionArgs) {
	await ensureUserNotLoggedIn(request);
	return null;
}

export default function Index() {
	return (
		<div className="min-h-screen bg-sky-600 flex flex-col items-center justify-center gap-y-6">
			<img
				src={`${BASE_PATH}logo.svg`}
				alt="iOSDC Japan 2025"
				className="w-64 h-64"
			/>
			<div className="text-center">
				<div className="font-bold text-sky-50 flex flex-col gap-y-2">
					<div className="text-5xl">SWIFT CODE BATTLE</div>
				</div>
			</div>
			<div className="mx-2">
				<BorderedContainer>
					<p className="text-gray-900 max-w-prose">
						Swift コードバトルは指示された動作をする Swift
						コードをより短く書けた方が勝ち、という 1 対 1
						の対戦コンテンツです。9/6
						に実施された予選を勝ち抜いたプレイヤーによるトーナメント形式での
						コードバトルを 9/19 (金) day0
						に実施します。ここでは短いコードが正義です！
						可読性も保守性も放り投げた、イベントならではのコードをお楽しみください！
					</p>
				</BorderedContainer>
			</div>
			<div>
				<NavigateLink to="/login">ログイン</NavigateLink>
			</div>
		</div>
	);
}
