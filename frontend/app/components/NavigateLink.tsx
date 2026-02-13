import { Link } from "wouter";

export default function NavigateLink({
	to,
	children,
}: {
	to: string;
	children: React.ReactNode;
}) {
	return (
		<Link
			to={to}
			className="text-lg text-white bg-sky-600 px-4 py-2 border-2 border-sky-50 rounded-sm transition duration-300 hover:bg-sky-500 focus:ring-3 focus:ring-sky-400 focus:outline-hidden"
		>
			{children}
		</Link>
	);
}
