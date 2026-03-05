type Props = {
  sec: number;
};

export default function LeftTime({ sec }: Props) {
  const s = sec % 60;
  const m = Math.floor(sec / 60) % 60;
  const h = Math.floor(sec / 3600) % 24;
  const d = Math.floor(sec / 86400);

  let leftTime = "";
  if (d > 0 || h > 0) {
    // 1d 2h 3m 4s
    leftTime = [
      d > 0 ? `${d}d` : "",
      h > 0 ? `${h}h` : "",
      m > 0 ? `${m}m` : "",
      `${s}s`,
    ].join(" ");
  } else {
    // 03:04
    leftTime = `${m.toString().padStart(2, "0")}:${s.toString().padStart(2, "0")}`;
  }

  return <div className="text-2xl md:text-3xl">{leftTime}</div>;
}
