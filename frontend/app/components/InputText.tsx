import React from "react";

type InputProps = React.InputHTMLAttributes<HTMLInputElement>;

export default function InputText(props: InputProps) {
  return (
    <input
      {...props}
      className="p-2 block w-full border border-brand-600 rounded-md transition duration-300 focus:ring-3 focus:ring-brand-400 focus:outline-hidden"
    />
  );
}
