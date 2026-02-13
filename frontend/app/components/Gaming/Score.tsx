import { useEffect, useState } from "react";

type Props = {
	status: string | null;
	score: number | null;
};

export default function Score({ status, score }: Props) {
	const [randomScore, setRandomScore] = useState<number | null>(null);

	useEffect(() => {
		if (status !== "running") {
			return;
		}

		const intervalId = setInterval(() => {
			const maxValue = Math.pow(10, String(score ?? 100).length) - 1;
			const minValue = Math.pow(10, String(score ?? 100).length - 1);
			const randomValue =
				Math.floor(Math.random() * (maxValue - minValue + 1)) + minValue;
			setRandomScore(randomValue);
		}, 50);

		return () => {
			clearInterval(intervalId);
		};
	}, [status, score]);

	const displayScore = status === "running" ? randomScore : score;

	return (
		<span className={status === "running" ? "animate-pulse" : ""}>
			{displayScore}
		</span>
	);
}
