import React, { useEffect } from "react";

export interface ModalProps {
  open: boolean;
  title: string;
  onClose: () => void;
  children: React.ReactNode;
}

export function Modal({ open, title, onClose, children }: ModalProps): JSX.Element | null {
  useEffect(() => {
    if (!open) return;
    const handler = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [open, onClose]);

  if (!open) return null;
  return (
    <div className="gc-modal-backdrop" onClick={onClose}>
      <div className="gc-modal" onClick={(e) => e.stopPropagation()}>
        <h2>{title}</h2>
        <div className="gc-modal-body">{children}</div>
      </div>
    </div>
  );
}
