import React from "react";

export interface ButtonProps {
  label: string;
  onClick?: () => void;
  disabled?: boolean;
}

export function Button({ label, onClick, disabled }: ButtonProps): JSX.Element {
  return (
    <button onClick={onClick} disabled={disabled} className="gc-btn">
      {label}
    </button>
  );
}
