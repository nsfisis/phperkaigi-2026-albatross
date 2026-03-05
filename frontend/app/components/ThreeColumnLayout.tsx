import React from "react";

type Props = {
  children: React.ReactNode;
};

export default function ThreeColumnLayout({ children }: Props) {
  return (
    <div className="grow grid grid-cols-1 md:grid-cols-3 md:divide-x divide-gray-300">
      {children}
    </div>
  );
}
