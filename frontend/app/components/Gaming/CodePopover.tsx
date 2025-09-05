import { Popover } from "@base-ui-components/react/popover";
import { faCode, faXmark } from "@fortawesome/free-solid-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { calcCodeSize } from "../../states/play";
import type { SupportedLanguage } from "../../types/SupportedLanguage";
import BorderedContainer from "../BorderedContainer";
import CodeBlock from "../Gaming/CodeBlock";

type Props = {
	code: string;
	language: SupportedLanguage;
};

export default function CodePopover({ code, language }: Props) {
	const codeSize = calcCodeSize(code, language);

	return (
		<Popover.Root>
			<Popover.Trigger>
				<FontAwesomeIcon icon={faCode} fixedWidth />
			</Popover.Trigger>
			<Popover.Portal>
				<Popover.Positioner>
					<Popover.Popup>
						<BorderedContainer className="grow flex flex-col gap-4">
							<div className="flex flex-row gap-2 items-center">
								<div className="grow font-semibold text-lg">
									コードサイズ: {codeSize}
								</div>
								<Popover.Close className="p-1 bg-gray-50 border-1 border-gray-300 rounded-sm">
									<FontAwesomeIcon
										icon={faXmark}
										fixedWidth
										className="text-gray-500"
									/>
								</Popover.Close>
							</div>
							<CodeBlock code={code} language={language} />
						</BorderedContainer>
					</Popover.Popup>
				</Popover.Positioner>
			</Popover.Portal>
		</Popover.Root>
	);
}
