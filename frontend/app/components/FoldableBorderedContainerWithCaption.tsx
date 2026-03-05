import { faChevronDown, faChevronUp } from "@fortawesome/free-solid-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import React, { useState } from "react";
import BorderedContainer from "./BorderedContainer";

type Props = {
  caption: string;
  children: React.ReactNode;
};

export default function FoldableBorderedContainerWithCaption({
  caption,
  children,
}: Props) {
  const [isOpen, setIsOpen] = useState(true);

  const handleToggle = () => {
    setIsOpen((prev) => !prev);
  };

  return (
    <BorderedContainer>
      <div className="flex flex-col gap-4">
        <div className="flex items-center">
          <div className="flex-1 text-center">
            <h2 className="text-lg font-semibold">{caption}</h2>
          </div>
          <div className="flex-shrink-0">
            <button
              onClick={handleToggle}
              className="p-1 bg-gray-50 border-1 border-gray-300 rounded-sm"
            >
              <FontAwesomeIcon
                icon={isOpen ? faChevronUp : faChevronDown}
                fixedWidth
                className="text-gray-500"
              />
            </button>
          </div>
        </div>
        <div className={isOpen ? "" : "hidden"}>{children}</div>
      </div>
    </BorderedContainer>
  );
}
