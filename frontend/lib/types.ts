export interface Coordinate {
  lat: number;
  lng: number;
}

export interface RouteRequest {
  start: Coordinate;
  stops: Coordinate[];
  end: Coordinate;
}

export type StopRole = "start" | "stop" | "end";

export interface OrderedStop {
  index: number;
  position: number;
  coordinate: Coordinate;
  role: StopRole;
  h3_cell: string;
}

export type TravelMode = "car" | "motorcycle" | "bike";

export interface ModeEstimate {
  mode: TravelMode;
  distance_meters: number;
  duration_seconds: number;
}

export interface H3Cluster {
  cell: string;
  center: Coordinate;
  boundary: number[][];
  stop_indexes: number[];
  count: number;
}

export interface Place {
  display_name: string;
  name: string;
  coordinate: Coordinate;
  category: string;
}

export interface RouteMetrics {
  distance_meters: number;
  duration_seconds: number;
}

export interface OptimizationResult {
  original_order: number[];
  optimized_order: number[];
  ordered_stops: OrderedStop[];
  original_metrics: RouteMetrics;
  optimized_metrics: RouteMetrics;
  distance_saved_meters: number;
  time_saved_seconds: number;
  efficiency_percent: number;
  optimization_millis: number;
  stop_count: number;
  geometry: number[][];
  original_geometry: number[][];
  duration_matrix: number[][];
  distance_matrix: number[][];
  mode_estimates: ModeEstimate[];
  h3_resolution: number;
  h3_clusters: H3Cluster[];
}

export interface ApiError {
  error: string;
}
