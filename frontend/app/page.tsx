"use client";

import dynamic from "next/dynamic";
import { useCallback, useRef, useState } from "react";
import Header, { Status } from "@/components/Header";
import ResultsPanel from "@/components/ResultsPanel";
import RouteSummary from "@/components/RouteSummary";
import SearchBox from "@/components/SearchBox";
import FloatingStats from "@/components/FloatingStats";
import { ToastStack, useToasts } from "@/components/Toasts";
import type { RouteState } from "@/components/MapView";
import { OptimizeError, optimizeRoute } from "@/lib/api";
import { formatDuration } from "@/lib/format";
import type { Coordinate, OptimizationResult, Place } from "@/lib/types";

const MapView = dynamic(() => import("@/components/MapView"), {
  ssr: false,
  loading: () => <div className="map-root" />,
});

export default function DispatchPage() {
  const [route, setRoute] = useState<RouteState>({ start: null, stops: [], end: null });
  const [finishing, setFinishing] = useState(false);
  const [result, setResult] = useState<OptimizationResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [status, setStatus] = useState<{ kind: Status; text: string }>({ kind: "ready", text: "Ready" });
  const [showOptimized, setShowOptimized] = useState(true);
  const [showOriginal, setShowOriginal] = useState(true);
  const [showHexGrid, setShowHexGrid] = useState(true);
  const fitRef = useRef<() => void>(() => {});
  const { toasts, notify, dismiss } = useToasts();

  const placeCoordinate = useCallback(
    (coord: Coordinate) => {
      setRoute((prev) => {
        if (!prev.start) {
          notify("info", "Start set", "Add delivery stops next.");
          return { ...prev, start: coord };
        }
        if (finishing) {
          setFinishing(false);
          notify("success", "Route complete", `${prev.stops.length} stop(s) ready to optimize.`);
          return { ...prev, end: coord };
        }
        if (!prev.end) {
          return { ...prev, stops: [...prev.stops, coord] };
        }
        return prev;
      });
    },
    [finishing, notify],
  );

  const handleSearchSelect = (place: Place) => {
    placeCoordinate(place.coordinate);
    notify("info", "Added from search", place.name);
  };

  const clearRoute = () => {
    setRoute({ start: null, stops: [], end: null });
    setFinishing(false);
    setResult(null);
    setStatus({ kind: "ready", text: "Ready" });
    notify("info", "Cleared", "Route reset. Click the map to begin.");
  };

  const startFinish = () => {
    if (!route.start || route.stops.length < 1) {
      notify("error", "Not ready", "Set a start and at least one stop first.");
      return;
    }
    setFinishing(true);
    notify("info", "Finish mode", "Click the map once to place the end location.");
  };

  const optimize = async () => {
    if (!route.start || route.stops.length < 1 || !route.end) {
      notify("error", "Route incomplete", "Set a start, at least one stop, and an end.");
      return;
    }

    setLoading(true);
    setStatus({ kind: "busy", text: "Optimizing…" });

    try {
      const data = await optimizeRoute({ start: route.start, stops: route.stops, end: route.end });
      setResult(data);
      setStatus({ kind: "ready", text: "Optimized" });
      notify("success", "Route optimized", `Saved ${formatDuration(data.time_saved_seconds)}.`);
      setTimeout(() => fitRef.current(), 50);
    } catch (err) {
      setStatus({ kind: "error", text: "Failed" });
      const message = err instanceof OptimizeError ? err.message : "Unexpected error. Please try again.";
      notify("error", "Optimization failed", message);
    } finally {
      setLoading(false);
    }
  };

  const routeReady = Boolean(route.start && route.stops.length >= 1 && route.end);

  return (
    <div className="app-shell">
      <Header status={status.kind} statusText={status.text} />
      <div className="body-grid">
        <aside className="sidebar" aria-label="Route controls">
          <div className="sidebar__section">
            <div className="section-title">Find a location</div>
            <SearchBox onSelect={handleSearchSelect} />
            <p className="hint" style={{ marginTop: 10 }}>
              Search a place, or click the map: first click sets the <strong>start</strong>, then add{" "}
              <strong>stops</strong>. Press <strong>Finish</strong>, then click once more for the <strong>end</strong>.
            </p>
          </div>

          <div className="sidebar__section">
            <div className="section-title">Current route</div>
            <RouteSummary route={route} finishing={finishing} />
          </div>

          <div className="sidebar__section">
            <div className="actions">
              <button className="btn btn--ghost btn--block" onClick={startFinish} type="button">
                Finish Route
              </button>
              <button
                className="btn btn--primary btn--block"
                onClick={optimize}
                type="button"
                disabled={loading || !routeReady}
              >
                {loading && <span className="btn__spinner" />}
                <span>Optimize Route</span>
              </button>
              <button className="btn btn--danger btn--block" onClick={clearRoute} type="button">
                Clear Route
              </button>
            </div>
          </div>

          {result && (
            <div className="sidebar__section" aria-live="polite">
              <ResultsPanel result={result} />
            </div>
          )}
        </aside>

        <main className="map-panel">
          <div className="map-toolbar">
            <div className="toolbar-group">
              <button
                className={`chip ${showOptimized ? "chip--active" : ""}`}
                onClick={() => setShowOptimized((v) => !v)}
                type="button"
              >
                <span className="chip__dot chip__dot--green" /> Optimized
              </button>
              <button
                className={`chip ${showOriginal ? "chip--active" : ""}`}
                onClick={() => setShowOriginal((v) => !v)}
                type="button"
              >
                <span className="chip__dot chip__dot--gray" /> Original
              </button>
              <button
                className={`chip ${showHexGrid ? "chip--active" : ""}`}
                onClick={() => setShowHexGrid((v) => !v)}
                type="button"
              >
                <span className="chip__dot chip__dot--hex" /> H3 zones
              </button>
            </div>
          </div>

          {result && <FloatingStats result={result} />}

          <MapView
            route={route}
            optimizedGeometry={result?.geometry ?? null}
            originalGeometry={result?.original_geometry ?? null}
            clusters={result?.h3_clusters ?? []}
            showOptimized={showOptimized}
            showOriginal={showOriginal}
            showHexGrid={showHexGrid}
            onMapClick={placeCoordinate}
            onDragStart={(c) => setRoute((prev) => ({ ...prev, start: c }))}
            onDragStop={(i, c) =>
              setRoute((prev) => ({ ...prev, stops: prev.stops.map((s, idx) => (idx === i ? c : s)) }))
            }
            onDragEnd={(c) => setRoute((prev) => ({ ...prev, end: c }))}
            registerFit={(fit) => {
              fitRef.current = fit;
            }}
          />

          <div className="legend" aria-hidden="true">
            <div className="legend__title">Legend</div>
            <div className="legend__row">
              <span className="legend__swatch legend__swatch--green" /> Start
            </div>
            <div className="legend__row">
              <span className="legend__swatch legend__swatch--blue" /> Delivery stop
            </div>
            <div className="legend__row">
              <span className="legend__swatch legend__swatch--red" /> End
            </div>
            <div className="legend__row">
              <span className="legend__swatch legend__swatch--line" /> Optimized route
            </div>
          </div>
        </main>
      </div>

      <ToastStack toasts={toasts} onDismiss={dismiss} />
    </div>
  );
}
