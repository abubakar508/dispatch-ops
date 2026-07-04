"use client";

import type { OptimizationResult } from "@/lib/types";
import { formatCoordinate, formatDistance, formatDuration } from "@/lib/format";

const badgeLabel = (role: string, position: number) => {
  if (role === "start") return "A";
  if (role === "end") return "B";
  return String(position);
};

const titleFor = (role: string, position: number) => {
  if (role === "start") return "Start";
  if (role === "end") return "End";
  return `Stop ${position}`;
};

export default function ResultsPanel({ result }: { result: OptimizationResult }) {
  return (
    <>
      <div className="section-title">Results</div>

      <div className="stats-grid">
        <div className="stat-card">
          <div className="stat-card__value">{result.stop_count}</div>
          <div className="stat-card__label">Delivery stops</div>
        </div>
        <div className="stat-card">
          <div className="stat-card__value">{formatDistance(result.optimized_metrics.distance_meters)}</div>
          <div className="stat-card__label">Total distance</div>
        </div>
        <div className="stat-card">
          <div className="stat-card__value">{formatDuration(result.optimized_metrics.duration_seconds)}</div>
          <div className="stat-card__label">Driving time</div>
        </div>
        <div className="stat-card">
          <div className="stat-card__value">{result.optimization_millis} ms</div>
          <div className="stat-card__label">Optimization time</div>
        </div>
        <div className="stat-card stat-card--accent">
          <div className="stat-card__value">{formatDuration(result.time_saved_seconds)}</div>
          <div className="stat-card__label">Time saved</div>
        </div>
        <div className="stat-card stat-card--accent">
          <div className="stat-card__value">{formatDistance(result.distance_saved_meters)}</div>
          <div className="stat-card__label">Distance saved</div>
        </div>
        <div className="stat-card stat-card--accent">
          <div className="stat-card__value">{Math.round(result.efficiency_percent)}%</div>
          <div className="stat-card__label">Route efficiency</div>
        </div>
      </div>

      <div className="section-title" style={{ marginTop: 18 }}>
        Optimized stop order
      </div>
      <ul className="route-list">
        {result.ordered_stops.map((stop) => (
          <li className="route-item" key={`${stop.index}-${stop.position}`}>
            <div className={`route-badge route-badge--${stop.role}`}>{badgeLabel(stop.role, stop.position)}</div>
            <div className="route-item__body">
              <div className="route-item__title">{titleFor(stop.role, stop.position)}</div>
              <div className="route-item__meta">{formatCoordinate(stop.coordinate)}</div>
              {stop.h3_cell && <div className="route-item__cell">H3 · {stop.h3_cell}</div>}
            </div>
          </li>
        ))}
      </ul>
    </>
  );
}
