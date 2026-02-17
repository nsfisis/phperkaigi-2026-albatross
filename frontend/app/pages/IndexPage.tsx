import BorderedContainer from "../components/BorderedContainer";
import NavigateLink from "../components/NavigateLink";
import { APP_NAME, BASE_PATH } from "../config";
import { usePageTitle } from "../hooks/usePageTitle";

export default function IndexPage() {
	usePageTitle(APP_NAME);

	return (
		<div className="min-h-screen bg-gray-100 flex flex-col items-center justify-center gap-y-6">
			<img
				src={`${BASE_PATH}logo.svg`}
				alt="PHPerKaigi 2026"
				className="w-96 h-auto"
			/>
			<div className="text-center">
				<div className="font-bold text-transparent bg-clip-text bg-phperkaigi">
					<div className="text-6xl">Swift Code Battle</div>
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
