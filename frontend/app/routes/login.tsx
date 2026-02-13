import type {
	ActionFunctionArgs,
	LoaderFunctionArgs,
	MetaFunction,
} from "react-router";
import { Form, data, useActionData } from "react-router";
import { ensureUserNotLoggedIn, login } from "../.server/auth";
import BorderedContainer from "../components/BorderedContainer";
import InputText from "../components/InputText";
import SubmitButton from "../components/SubmitButton";
import { APP_NAME } from "../config";

export const meta: MetaFunction = () => [{ title: `Login | ${APP_NAME}` }];

export async function loader({ request }: LoaderFunctionArgs) {
	return await ensureUserNotLoggedIn(request);
}

export async function action({ request }: ActionFunctionArgs) {
	const formData = await request.clone().formData();
	const username = String(formData.get("username"));
	const password = String(formData.get("password"));
	if (username === "" || password === "") {
		return data(
			{
				message: "ユーザー名またはパスワードが誤っています",
				errors: {
					username:
						username === "" ? "ユーザー名を入力してください" : undefined,
					password:
						password === "" ? "パスワードを入力してください" : undefined,
				},
			},
			{ status: 400 },
		);
	}

	try {
		await login(request);
	} catch (error) {
		if (error instanceof Error) {
			return data(
				{
					message: error.message,
					errors: {
						username: undefined,
						password: undefined,
					},
				},
				{ status: 400 },
			);
		} else {
			throw error;
		}
	}
	return null;
}

export default function Login() {
	const loginErrors = useActionData<typeof action>();

	return (
		<div className="min-h-screen bg-gray-100 flex items-center justify-center">
			<div className="mx-2">
				<BorderedContainer>
					<Form method="post" className="w-full max-w-sm p-2">
						<h2 className="text-2xl mb-6 text-center">
							fortee アカウントでログイン
						</h2>
						{loginErrors?.message && (
							<p className="text-sky-500 text-sm mb-4">{loginErrors.message}</p>
						)}
						<div className="mb-4 flex flex-col gap-1">
							<label
								htmlFor="username"
								className="block text-sm font-medium text-gray-700"
							>
								ユーザー名
							</label>
							<InputText type="text" name="username" id="username" required />
							{loginErrors?.errors?.username && (
								<p className="text-red-500 text-sm">
									{loginErrors.errors.username}
								</p>
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
							{loginErrors?.errors?.password && (
								<p className="text-red-500 text-sm">
									{loginErrors.errors.password}
								</p>
							)}
						</div>
						<div className="flex justify-center">
							<SubmitButton type="submit">ログイン</SubmitButton>
						</div>
					</Form>
				</BorderedContainer>
			</div>
		</div>
	);
}
