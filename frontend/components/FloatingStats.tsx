"use client";

import { useState } from "react";
import type { OptimizationResult, TravelMode } from "@/lib/types";
import { formatDistance, formatDuration } from "@/lib/format";
import { MODE_LABELS, MODE_ORDER, ModeIcon } from "@/lib/modes";

export default function FloatingStats({ result }: { result: OptimizationResult }) {
  const [mode, setMode] = useState<TravelMode>("car");

  const estimate = result.mode_estimates.find((e) => e.mode === mode) ?? {
    distance_meters: result.optimized_metrics.distance_meters,
    duration_seconds: result.optimized_metrics.duration_seconds,
    mode,
  };

  return (
    <div className="floating-stats" role="region" aria-label="Route statistics">
      <div className="floating-stats__header">
        <span className="floating-stats__title">Route summary</span>
        <span className="floating-stats__pill">{Math.round(result.efficiency_percent)}% faster</span>
      </div>

      <div className="mode-tabs" role="tablist" aria-label="Travel mode">
        {MODE_ORDER.map((m) => (
          <button
            key={m}
            role="tab"
            aria-selected={mode === m}
            className={`mode-tab ${mode === m ? "mode-tab--active" : ""}`}
            onClick={() => setMode(m)}
            type="button"
          >
            <ModeIcon mode={m} />
            <span>{MODE_LABELS[m]}</span>
          </button>
        ))}
      </div>

      <div className="floating-stats__grid">
        <div className="floating-stat">
          <div className="floating-stat__value">{formatDuration(estimate.duration_seconds)}</div>
          <div className="floating-stat__label">Est. time · {MODE_LABELS[mode]}</div>
        </div>
        <div className="floating-stat">
          <div className="floating-stat__value">{formatDistance(estimate.distance_meters)}</div>
          <div className="floating-stat__label">Distance</div>
        </div>
        <div className="floating-stat">
          <div className="floating-stat__value">{result.stop_count}</div>
          <div className="floating-stat__label">Stops</div>
        </div>
        <div className="floating-stat">
          <div className="floating-stat__value">{result.h3_clusters.length}</div>
          <div className="floating-stat__label">H3 zones</div>
        </div>
      </div>

      <div className="floating-stats__foot">
        <span>Saved {formatDuration(result.time_saved_seconds)}</span>
        <span>{result.optimization_millis} ms solve</span>
      </div>
    </div>
  );
}
