import { useAtomValue } from "jotai";
import React, { useRef, useState } from "react";
import { Link } from "wouter";
import type { components } from "../../api/schema";
import {
  calcCodeSize,
  gamingLeftTimeSecondsAtom,
  scoreAtom,
  statusAtom,
} from "../../states/play";
import type { PlayerProfile } from "../../types/PlayerProfile";
import type { SupportedLanguage } from "../../types/SupportedLanguage";
import BorderedContainer from "../BorderedContainer";
import CodePopover from "../Gaming/CodePopover";
import DataTable, {
  DataTableCell,
  formatUnixTimestamp,
} from "../Gaming/DataTable";
import LeftTime from "../Gaming/LeftTime";
import ProblemColumn from "../Gaming/ProblemColumn";
import SubmitButton from "../SubmitButton";
import SubmitStatusLabel from "../SubmitStatusLabel";
import ThreeColumnLayout from "../ThreeColumnLayout";
import TitledColumn from "../TitledColumn";
import UserIcon from "../UserIcon";

type Submission = components["schemas"]["Submission"];

type Props = {
  gameDisplayName: string;
  playerProfile: PlayerProfile;
  problemTitle: string;
  problemDescription: string;
  problemLanguage: SupportedLanguage;
  sampleCode: string;
  initialCode: string;
  onCodeChange: (code: string) => void;
  onCodeSubmit: (code: string) => void;
  isFinished: boolean;
  submissions: Submission[];
};

export default function GolfPlayAppGaming({
  gameDisplayName,
  playerProfile,
  problemTitle,
  problemDescription,
  problemLanguage,
  sampleCode,
  initialCode,
  onCodeChange,
  onCodeSubmit,
  isFinished,
  submissions,
}: Props) {
  const leftTimeSeconds = useAtomValue(gamingLeftTimeSecondsAtom);
  const score = useAtomValue(scoreAtom);
  const status = useAtomValue(statusAtom);

  const [codeSize, setCodeSize] = useState(
    calcCodeSize(initialCode, problemLanguage),
  );
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const handleTextChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setCodeSize(calcCodeSize(e.target.value, problemLanguage));
    if (!isFinished) {
      onCodeChange(e.target.value);
    }
  };

  const handleSubmitButtonClick = () => {
    if (textareaRef.current && !isFinished) {
      onCodeSubmit(textareaRef.current.value);
    }
  };

  return (
    <div className="min-h-screen bg-gray-100 flex flex-col">
      <div className="text-white bg-brand-600 flex flex-row justify-between px-4 py-2">
        <div className="font-bold">
          <div className="text-gray-100">{gameDisplayName}</div>
          {isFinished ? (
            <div className="text-2xl md:text-3xl">終了</div>
          ) : leftTimeSeconds === null ? (
            <div className="text-2xl md:text-3xl">未開始</div>
          ) : (
            <LeftTime sec={leftTimeSeconds} />
          )}
        </div>
        <Link to={"/dashboard"}>
          <div className="flex gap-6 items-center font-bold">
            <div className="text-2xl md:text-6xl">{score}</div>
            <div className="hidden md:block text-4xl">
              {playerProfile.displayName}
            </div>
            {playerProfile.iconPath && (
              <UserIcon
                iconPath={playerProfile.iconPath}
                displayName={playerProfile.displayName}
                className="w-12 h-12 my-auto"
              />
            )}
          </div>
        </Link>
      </div>
      <ThreeColumnLayout>
        <ProblemColumn
          title={problemTitle}
          description={problemDescription}
          language={problemLanguage}
          sampleCode={sampleCode}
        />
        <TitledColumn title="ソースコード">
          <BorderedContainer className="grow flex flex-col gap-4">
            <div className="flex flex-row gap-2 items-center">
              <div className="grow font-semibold text-lg">
                コードサイズ: {codeSize}
              </div>
              <SubmitButton
                onClick={handleSubmitButtonClick}
                disabled={isFinished}
              >
                提出
              </SubmitButton>
            </div>
            <textarea
              ref={textareaRef}
              defaultValue={initialCode}
              onChange={handleTextChange}
              className="grow resize-none h-full w-full p-2 bg-gray-50 rounded-lg border border-gray-300 focus:outline-hidden focus:ring-2 focus:ring-gray-400 transition duration-300"
              rows={10}
            />
          </BorderedContainer>
        </TitledColumn>
        <TitledColumn title="提出結果">
          <DataTable headers={["ステータス", "スコア", "提出時刻", "コード"]}>
            {submissions.length === 0 ? (
              <tr>
                <DataTableCell>
                  <SubmitStatusLabel status={status} />
                </DataTableCell>
                <DataTableCell>-</DataTableCell>
                <DataTableCell>-</DataTableCell>
                <DataTableCell>-</DataTableCell>
              </tr>
            ) : (
              submissions.map((s) => (
                <tr key={s.submission_id}>
                  <DataTableCell>
                    <SubmitStatusLabel status={s.status} />
                  </DataTableCell>
                  <DataTableCell>{s.code_size}</DataTableCell>
                  <DataTableCell>
                    {formatUnixTimestamp(s.created_at)}
                  </DataTableCell>
                  <DataTableCell>
                    <CodePopover code={s.code} language={problemLanguage} />
                  </DataTableCell>
                </tr>
              ))
            )}
          </DataTable>
        </TitledColumn>
      </ThreeColumnLayout>
    </div>
  );
}
