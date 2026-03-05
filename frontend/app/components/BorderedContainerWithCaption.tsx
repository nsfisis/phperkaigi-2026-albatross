import React from "react";
import BorderedContainer from "./BorderedContainer";

type Props = {
  caption: string;
  children: React.ReactNode;
};

export default function BorderedContainerWithCaption({
  caption,
  children,
}: Props) {
  return (
    <BorderedContainer>
      <div className="flex flex-col gap-4">
        <h2 className="text-center text-lg font-semibold">{caption}</h2>
        {children}
      </div>
    </BorderedContainer>
  );
}
