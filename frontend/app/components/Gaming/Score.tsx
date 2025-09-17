import { useEffect, useState } from "react";

type Props = {
	status: string | null;
	score: number | null;
};

export default function Score({ status, score }: Props) {
	const [displayScore, setDisplayScore] = useState(score);

	useEffect(() => {
		let intervalId = null;

		if (status === "running") {
			intervalId = setInterval(() => {
				const maxValue = Math.pow(10, String(score ?? 100).length) - 1;
				const minValue = Math.pow(10, String(score ?? 100).length - 1);
				const randomValue =
					Math.floor(Math.random() * (maxValue - minValue + 1)) + minValue;
				setDisplayScore(randomValue);
			}, 50);
		} else {
			setDisplayScore(score);
		}

		return () => {
			if (intervalId) {
				clearInterval(intervalId);
			}
		};
	}, [status, score]);

	return (
		<span className={status === "running" ? "animate-pulse" : ""}>
			{displayScore}
		</span>
	);
}
