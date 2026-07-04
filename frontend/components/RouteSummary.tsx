"use client";

import type { RouteState } from "./MapView";
import { formatCoordinate } from "@/lib/format";

interface RouteSummaryProps {
  route: RouteState;
  finishing: boolean;
}

export default function RouteSummary({ route, finishing }: RouteSummaryProps) {
  const empty = !route.start && route.stops.length === 0 && !route.end;

  if (empty) {
    return (
      <div className="empty-state">
        <div className="empty-state__icon">📍</div>
        <div className="empty-state__title">No route yet</div>
        <div className="empty-state__text">Click the map to set a start, then add delivery stops.</div>
      </div>
    );
  }

  return (
    <ul className="route-list">
      {route.start && (
        <li className="route-item">
          <div className="route-badge route-badge--start">A</div>
          <div className="route-item__body">
            <div className="route-item__title">Start</div>
            <div className="route-item__meta">{formatCoordinate(route.start)}</div>
          </div>
        </li>
      )}
      {route.stops.map((stop, i) => (
        <li className="route-item" key={i}>
          <div className="route-badge route-badge--stop">{i + 1}</div>
          <div className="route-item__body">
            <div className="route-item__title">Stop {i + 1}</div>
            <div className="route-item__meta">{formatCoordinate(stop)}</div>
          </div>
        </li>
      ))}
      {route.end ? (
        <li className="route-item">
          <div className="route-badge route-badge--end">B</div>
          <div className="route-item__body">
            <div className="route-item__title">End</div>
            <div className="route-item__meta">{formatCoordinate(route.end)}</div>
          </div>
        </li>
      ) : (
        finishing && <li className="hint" style={{ padding: "8px 0" }}>Click the map to place the end…</li>
      )}
    </ul>
  );
}
