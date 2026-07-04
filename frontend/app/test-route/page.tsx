"use client";

import { useState } from "react";
import Header from "@/components/Header";
import { OptimizeError, optimizeRoute } from "@/lib/api";
import { formatDistance, formatDuration } from "@/lib/format";
import { MODE_LABELS } from "@/lib/modes";
import type { Coordinate, OptimizationResult } from "@/lib/types";

interface Row {
  lat: string;
  lng: string;
}

const emptyRow: Row = { lat: "", lng: "" };

export default function TestRoutePage() {
  const [start, setStart] = useState<Row>({ ...emptyRow });
  const [end, setEnd] = useState<Row>({ ...emptyRow });
  const [stops, setStops] = useState<Row[]>([{ ...emptyRow }, { ...emptyRow }]);
  const [result, setResult] = useState<OptimizationResult | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const parse = (row: Row): Coordinate => ({ lat: parseFloat(row.lat), lng: parseFloat(row.lng) });

  const addStop = () => setStops((prev) => [...prev, { ...emptyRow }]);
  const removeStop = (index: number) => setStops((prev) => prev.filter((_, i) => i !== index));
  const updateStop = (index: number, field: keyof Row, value: string) =>
    setStops((prev) => prev.map((r, i) => (i === index ? { ...r, [field]: value } : r)));

  const seedSample = () => {
    setStart({ lat: "-1.2864", lng: "36.8172" });
    setEnd({ lat: "-1.2999", lng: "36.7820" });
    setStops([
      { lat: "-1.2683", lng: "36.8109" },
      { lat: "-1.2635", lng: "36.8506" },
      { lat: "-1.3182", lng: "36.8283" },
      { lat: "-1.2921", lng: "36.8219" },
    ]);
  };

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      const data = await optimizeRoute({
        start: parse(start),
        end: parse(end),
        stops: stops.map(parse),
      });
      setResult(data);
    } catch (err) {
      setResult(null);
      setError(err instanceof OptimizeError ? err.message : "Network error. Could not reach the server.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="app-shell">
      <Header />
      <div className="test-layout">
        <section className="test-panel">
          <h1 className="page-title">Test Route</h1>
          <p className="page-subtitle">
            Engineering tool for verifying the optimizer. Enter raw coordinates, then inspect the OSRM matrix and solver
            output.
          </p>

          <form onSubmit={submit} noValidate>
            <div className="card">
              <div className="card__title">Start</div>
              <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 6 }}>
                <input className="input" type="number" step="any" placeholder="Latitude" value={start.lat} onChange={(e) => setStart({ ...start, lat: e.target.value })} required />
                <input className="input" type="number" step="any" placeholder="Longitude" value={start.lng} onChange={(e) => setStart({ ...start, lng: e.target.value })} required />
              </div>
            </div>

            <div className="card">
              <div className="card__title">Delivery stops</div>
              {stops.map((row, i) => (
                <div className="stop-row" key={i}>
                  <div className="stop-row__num">{i + 1}</div>
                  <input className="input" type="number" step="any" placeholder="Latitude" value={row.lat} onChange={(e) => updateStop(i, "lat", e.target.value)} required />
                  <input className="input" type="number" step="any" placeholder="Longitude" value={row.lng} onChange={(e) => updateStop(i, "lng", e.target.value)} required />
                  <button type="button" className="remove-btn" aria-label="Remove stop" onClick={() => removeStop(i)}>
                    &times;
                  </button>
                </div>
              ))}
              <button type="button" className="btn btn--ghost btn--sm" style={{ marginTop: 6 }} onClick={addStop}>
                + Add Stop
              </button>
            </div>

            <div className="card">
              <div className="card__title">End</div>
              <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 6 }}>
                <input className="input" type="number" step="any" placeholder="Latitude" value={end.lat} onChange={(e) => setEnd({ ...end, lat: e.target.value })} required />
                <input className="input" type="number" step="any" placeholder="Longitude" value={end.lng} onChange={(e) => setEnd({ ...end, lng: e.target.value })} required />
              </div>
            </div>

            <div className="actions">
              <button className="btn btn--primary btn--block" type="submit" disabled={loading}>
                {loading && <span className="btn__spinner" />}
                <span>Optimize</span>
              </button>
              <button className="btn btn--ghost btn--block" type="button" onClick={seedSample}>
                Load Nairobi sample
              </button>
            </div>
          </form>
        </section>

        <section className="test-output">
          <h2 className="page-title" style={{ fontSize: 16, marginBottom: 14 }}>
            Solver output
          </h2>

          {error && <div className="error-box">{error}</div>}

          {!result && !error && (
            <div className="empty-state">
              <div className="empty-state__icon">🧪</div>
              <div className="empty-state__title">No run yet</div>
              <div className="empty-state__text">Enter coordinates or load the sample, then press Optimize.</div>
            </div>
          )}

          {result && <SolverOutput result={result} />}
        </section>
      </div>
    </div>
  );
}

function SolverOutput({ result }: { result: OptimizationResult }) {
  return (
    <>
      <div className="card">
        <div className="card__title">Summary</div>
        <div className="stats-grid">
          <div className="stat-card">
            <div className="stat-card__value">{result.stop_count}</div>
            <div className="stat-card__label">Stops</div>
          </div>
          <div className="stat-card">
            <div className="stat-card__value">{formatDistance(result.optimized_metrics.distance_meters)}</div>
            <div className="stat-card__label">Distance</div>
          </div>
          <div className="stat-card">
            <div className="stat-card__value">{formatDuration(result.optimized_metrics.duration_seconds)}</div>
            <div className="stat-card__label">Driving time</div>
          </div>
          <div className="stat-card">
            <div className="stat-card__value">{result.optimization_millis} ms</div>
            <div className="stat-card__label">Optimization duration</div>
          </div>
        </div>
      </div>

      <div className="card">
        <div className="card__title">Per-mode estimates</div>
        <div className="matrix-wrap">
          <table className="matrix-table">
            <thead>
              <tr>
                <th>Mode</th>
                <th>Distance</th>
                <th>Duration</th>
              </tr>
            </thead>
            <tbody>
              {result.mode_estimates.map((m) => (
                <tr key={m.mode}>
                  <th>{MODE_LABELS[m.mode]}</th>
                  <td>{formatDistance(m.distance_meters)}</td>
                  <td>{formatDuration(m.duration_seconds)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      <div className="card">
        <div className="card__title">
          H3 zones · resolution {result.h3_resolution} · {result.h3_clusters.length} cells
        </div>
        <div className="order-tags">
          {result.h3_clusters.map((c) => (
            <span className="order-tag order-tag--opt" key={c.cell}>
              {c.cell} ({c.count})
            </span>
          ))}
        </div>
      </div>

      <div className="card">
        <div className="card__title">Original order</div>
        <div className="order-tags">
          {result.original_order.map((n, i) => (
            <span className="order-tag" key={i}>
              {n}
            </span>
          ))}
        </div>
        <div className="card__title" style={{ marginTop: 14 }}>
          Optimized order (solver solution)
        </div>
        <div className="order-tags">
          {result.optimized_order.map((n, i) => (
            <span className="order-tag order-tag--opt" key={i}>
              {n}
            </span>
          ))}
        </div>
      </div>

      <div className="card">
        <div className="card__title">Raw OSRM duration matrix (seconds)</div>
        <div className="matrix-wrap">
          <table className="matrix-table">
            <thead>
              <tr>
                <th>#</th>
                {result.duration_matrix.map((_, i) => (
                  <th key={i}>{i}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {result.duration_matrix.map((row, i) => (
                <tr key={i}>
                  <th>{i}</th>
                  {row.map((cell, j) => (
                    <td key={j}>{Math.round(cell)}</td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      <div className="card">
        <div className="card__title">Optimized geometry (OSRM route, GeoJSON coordinates)</div>
        <pre className="mono">{JSON.stringify(result.geometry)}</pre>
      </div>
    </>
  );
}
