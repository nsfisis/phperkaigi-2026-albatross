import type { components } from "../api/schema";

type Props = {
	status: components["schemas"]["ExecutionStatus"];
};

export default function SubmitStatusLabel({ status }: Props) {
	switch (status) {
		case "none":
			return "提出待ち";
		case "running":
			return "実行中...";
		case "success":
			return "成功";
		case "wrong_answer":
			return "テスト失敗";
		case "timeout":
			return "時間切れ";
		case "compile_error":
			return "コンパイルエラー";
		case "runtime_error":
			return "実行時エラー";
		case "internal_error":
			return "！内部エラー！";
	}
}
