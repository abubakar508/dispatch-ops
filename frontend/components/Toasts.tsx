"use client";

import { useCallback, useState } from "react";

export type ToastKind = "success" | "error" | "info";

export interface Toast {
  id: number;
  kind: ToastKind;
  title: string;
  text: string;
}

export function useToasts() {
  const [toasts, setToasts] = useState<Toast[]>([]);

  const dismiss = useCallback((id: number) => {
    setToasts((current) => current.filter((t) => t.id !== id));
  }, []);

  const notify = useCallback(
    (kind: ToastKind, title: string, text: string) => {
      const id = Date.now() + Math.random();
      setToasts((current) => [...current, { id, kind, title, text }]);
      setTimeout(() => dismiss(id), 3800);
    },
    [dismiss],
  );

  return { toasts, notify, dismiss };
}

export function ToastStack({ toasts, onDismiss }: { toasts: Toast[]; onDismiss: (id: number) => void }) {
  return (
    <div className="toast-stack" aria-live="assertive">
      {toasts.map((toast) => (
        <div
          key={toast.id}
          className={`toast toast--${toast.kind}`}
          role="button"
          tabIndex={0}
          onClick={() => onDismiss(toast.id)}
          onKeyDown={(e) => e.key === "Enter" && onDismiss(toast.id)}
        >
          <div>
            <div className="toast__title">{toast.title}</div>
            <div className="toast__text">{toast.text}</div>
          </div>
        </div>
      ))}
    </div>
  );
}
