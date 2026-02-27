import React from "react";

type Props = {
	children: React.ReactNode;
	className?: string;
};

export default function BorderedContainer({ children, className }: Props) {
	return (
		<div
			className={`bg-white border-2 border-brand-600 rounded-xl p-4 ${className}`}
		>
			{children}
		</div>
	);
}
