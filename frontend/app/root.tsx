import { config } from "@fortawesome/fontawesome-svg-core";
import "@fortawesome/fontawesome-svg-core/styles.css";
import type { LinksFunction } from "react-router";
import { Links, Meta, Outlet, Scripts, ScrollRestoration } from "react-router";
import "./tailwind.css";
import "./shiki.css";
import { BASE_PATH } from "./config";

config.autoAddCss = false;

export const links: LinksFunction = () => [
	{ rel: "icon", href: `${BASE_PATH}code-battle/favicon.svg` },
];

export function Layout({ children }: { children: React.ReactNode }) {
	return (
		<html lang="ja">
			<head>
				<meta charSet="utf-8" />
				<meta name="viewport" content="width=device-width, initial-scale=1" />
				<Meta />
				<Links />
			</head>
			<body className="h-screen">
				{children}
				<ScrollRestoration />
				<Scripts />
				<script>console.log(`#Albatross!`)</script>
			</body>
		</html>
	);
}

export default function App() {
	return <Outlet />;
}
