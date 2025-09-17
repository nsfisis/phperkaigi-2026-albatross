import { faCopy } from "@fortawesome/free-solid-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { JSX, useLayoutEffect, useState } from "react";
import { type BundledLanguage, highlight } from "../../highlight";

type Props = {
	code: string;
	language: BundledLanguage;
};

export default function CodeBlock({ code, language }: Props) {
	const [nodes, setNodes] = useState<JSX.Element | null>(null);
	const [showCopied, setShowCopied] = useState(false);

	useLayoutEffect(() => {
		highlight(code, language).then(setNodes);
	}, [code, language]);

	const handleCopy = () => {
		navigator.clipboard.writeText(code).then(() => {
			setShowCopied(true);
			setTimeout(() => setShowCopied(false), 3000);
		});
	};

	return (
		<div className="relative">
			{code !== "" && (
				<button
					onClick={handleCopy}
					className="absolute top-2 right-2 z-10 px-2 py-1 bg-white border border-gray-300 rounded shadow-md hover:bg-gray-100 transition-colors"
					title="コードをコピーする"
				>
					<FontAwesomeIcon icon={faCopy} className="text-gray-600" />
					{showCopied && (
						<span className="ml-1 text-xs text-blue-600">Copied!</span>
					)}
				</button>
			)}
			<pre className="h-full w-full p-2 bg-gray-50 rounded-lg border border-gray-300 whitespace-pre-wrap break-words">
				{nodes === null ? <code>{code}</code> : nodes}
			</pre>
		</div>
	);
}
