type Props = {
  code: string;
};

export default function InlineCode({ code }: Props) {
  return (
    <code className="bg-gray-50 rounded-lg border border-gray-300 p-1">
      {code}
    </code>
  );
}
