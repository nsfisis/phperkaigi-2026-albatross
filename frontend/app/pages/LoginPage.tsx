import { type FormEvent, useState } from "react";
import { useLocation } from "wouter";
import BorderedContainer from "../components/BorderedContainer";
import InputText from "../components/InputText";
import SubmitButton from "../components/SubmitButton";
import { APP_NAME } from "../config";
import { useAuth } from "../hooks/useAuth";
import { usePageTitle } from "../hooks/usePageTitle";

export default function LoginPage() {
	usePageTitle(`Login | ${APP_NAME}`);

	const { login } = useAuth();
	const [, navigate] = useLocation();

	const [error, setError] = useState<string | null>(null);
	const [fieldErrors, setFieldErrors] = useState<{
		username?: string;
		password?: string;
	}>({});
	const [submitting, setSubmitting] = useState(false);

	async function handleSubmit(e: FormEvent<HTMLFormElement>) {
		e.preventDefault();
		const formData = new FormData(e.currentTarget);
		const username = String(formData.get("username"));
		const password = String(formData.get("password"));

		const errors: { username?: string; password?: string } = {};
		if (username === "") errors.username = "ユーザー名を入力してください";
		if (password === "") errors.password = "パスワードを入力してください";
		if (Object.keys(errors).length > 0) {
			setFieldErrors(errors);
			setError("ユーザー名またはパスワードが誤っています");
			return;
		}

		setSubmitting(true);
		setError(null);
		setFieldErrors({});

		try {
			await login(username, password);
			navigate("/dashboard");
		} catch (err) {
			setError(err instanceof Error ? err.message : "ログインに失敗しました");
		} finally {
			setSubmitting(false);
		}
	}

	return (
		<div className="min-h-screen bg-gray-100 flex items-center justify-center">
			<div className="mx-2">
				<BorderedContainer>
					<form onSubmit={handleSubmit} className="w-full max-w-sm p-2">
						<h2 className="text-2xl mb-6 text-center">
							fortee アカウントでログイン
						</h2>
						{error && <p className="text-sky-500 text-sm mb-4">{error}</p>}
						<div className="mb-4 flex flex-col gap-1">
							<label
								htmlFor="username"
								className="block text-sm font-medium text-gray-700"
							>
								ユーザー名
							</label>
							<InputText type="text" name="username" id="username" required />
							{fieldErrors.username && (
								<p className="text-red-500 text-sm">{fieldErrors.username}</p>
							)}
						</div>
						<div className="mb-6 flex flex-col gap-1">
							<label
								htmlFor="password"
								className="block text-sm font-medium text-gray-700"
							>
								パスワード
							</label>
							<InputText
								type="password"
								name="password"
								id="password"
								autoComplete="current-password"
								required
							/>
							{fieldErrors.password && (
								<p className="text-red-500 text-sm">{fieldErrors.password}</p>
							)}
						</div>
						<div className="flex justify-center">
							<SubmitButton type="submit" disabled={submitting}>
								{submitting ? "ログイン中..." : "ログイン"}
							</SubmitButton>
						</div>
					</form>
				</BorderedContainer>
			</div>
		</div>
	);
}
