import React from "react";

type Props = {
  children: React.ReactNode;
  title: React.ReactNode;
  className?: string;
};

export default function TitledColumn({ children, title, className }: Props) {
  return (
    <div className={`p-4 flex flex-col gap-4 ${className}`}>
      <div className="text-center text-xl font-bold">{title}</div>
      {children}
    </div>
  );
}
