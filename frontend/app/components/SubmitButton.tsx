import React from "react";

type ButtonProps = React.ButtonHTMLAttributes<HTMLButtonElement>;

export default function SubmitButton(props: ButtonProps) {
	return (
		<button
			{...props}
			className="text-lg text-white px-4 py-2 bg-brand-600 disabled:bg-gray-400 disabled:cursor-not-allowed rounded-sm transition duration-300 hover:bg-brand-500 focus:ring-3 focus:ring-brand-400 focus:outline-hidden"
		/>
	);
}
