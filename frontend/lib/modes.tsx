import type { TravelMode } from "./types";

export const MODE_ORDER: TravelMode[] = ["car", "motorcycle", "bike"];

export const MODE_LABELS: Record<TravelMode, string> = {
  car: "Car",
  motorcycle: "Motorcycle",
  bike: "Bike",
};

export function ModeIcon({ mode, size = 18 }: { mode: TravelMode; size?: number }) {
  const common = {
    width: size,
    height: size,
    viewBox: "0 0 24 24",
    fill: "none",
    stroke: "currentColor",
    strokeWidth: 1.8,
    strokeLinecap: "round" as const,
    strokeLinejoin: "round" as const,
  };

  if (mode === "car") {
    return (
      <svg {...common} aria-hidden="true">
        <path d="M5 13l1.5-4.5A2 2 0 0 1 8.4 7h7.2a2 2 0 0 1 1.9 1.5L19 13" />
        <path d="M4 17h16v-3.2a1 1 0 0 0-.6-.9L18 12H6l-1.4.9a1 1 0 0 0-.6.9V17z" />
        <circle cx="7.5" cy="17.5" r="1.6" />
        <circle cx="16.5" cy="17.5" r="1.6" />
      </svg>
    );
  }

  if (mode === "motorcycle") {
    return (
      <svg {...common} aria-hidden="true">
        <circle cx="5.5" cy="16.5" r="3.2" />
        <circle cx="18.5" cy="16.5" r="3.2" />
        <path d="M5.5 16.5l3-5h5l2 3" />
        <path d="M13.5 8h3l1.5 3" />
        <path d="M8.5 11.5H14" />
      </svg>
    );
  }

  return (
    <svg {...common} aria-hidden="true">
      <circle cx="5.5" cy="17" r="3.2" />
      <circle cx="18.5" cy="17" r="3.2" />
      <path d="M5.5 17l4-7h4" />
      <path d="M9.5 10l3 7" />
      <path d="M12.5 10H15" />
      <path d="M14 7l2 3" />
    </svg>
  );
}
