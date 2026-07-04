import type { ApiError, OptimizationResult, Place, RouteRequest } from "./types";

export async function searchPlaces(query: string, signal?: AbortSignal): Promise<Place[]> {
  const trimmed = query.trim();
  if (!trimmed) return [];

  const response = await fetch(`/api/search?q=${encodeURIComponent(trimmed)}&limit=6`, { signal });
  if (!response.ok) return [];

  const data = (await response.json()) as { results?: Place[] };
  return data.results ?? [];
}

export class OptimizeError extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.name = "OptimizeError";
    this.status = status;
  }
}

export async function optimizeRoute(request: RouteRequest): Promise<OptimizationResult> {
  let response: Response;
  try {
    response = await fetch("/api/optimize", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(request),
    });
  } catch {
    throw new OptimizeError("Could not reach the optimization service. Check your connection and try again.", 0);
  }

  const text = await response.text();

  if (!response.ok) {
    let message = "Something went wrong while optimizing this route.";
    try {
      const parsed = JSON.parse(text) as ApiError;
      if (parsed.error) message = parsed.error;
    } catch {
      /* keep default message */
    }
    throw new OptimizeError(message, response.status);
  }

  return JSON.parse(text) as OptimizationResult;
}
