"use client";

import { useEffect, useRef } from "react";
import type { Map as LeafletMap, Marker, Polygon, Polyline } from "leaflet";
import type { Coordinate, H3Cluster } from "@/lib/types";

const NAIROBI: [number, number] = [-1.2921, 36.8219];

export interface RouteState {
  start: Coordinate | null;
  stops: Coordinate[];
  end: Coordinate | null;
}

interface MapViewProps {
  route: RouteState;
  optimizedGeometry: number[][] | null;
  originalGeometry: number[][] | null;
  clusters: H3Cluster[];
  showOptimized: boolean;
  showOriginal: boolean;
  showHexGrid: boolean;
  onMapClick: (coord: Coordinate) => void;
  onDragStart: (coord: Coordinate) => void;
  onDragStop: (index: number, coord: Coordinate) => void;
  onDragEnd: (coord: Coordinate) => void;
  registerFit: (fit: () => void) => void;
}

const round = (n: number) => Math.round(n * 1e6) / 1e6;

export default function MapView(props: MapViewProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const mapRef = useRef<LeafletMap | null>(null);
  const leafletRef = useRef<typeof import("leaflet") | null>(null);
  const markersRef = useRef<Marker[]>([]);
  const optimizedRef = useRef<Polyline | null>(null);
  const originalRef = useRef<Polyline | null>(null);
  const hexRef = useRef<Polygon[]>([]);
  const propsRef = useRef(props);
  propsRef.current = props;

  useEffect(() => {
    let cancelled = false;

    import("leaflet").then((L) => {
      if (cancelled || !containerRef.current || mapRef.current) return;
      leafletRef.current = L;

      const map = L.map(containerRef.current, { zoomControl: false, attributionControl: true }).setView(NAIROBI, 12);
      L.tileLayer("https://{s}.basemaps.cartocdn.com/light_nolabels/{z}/{x}/{y}{r}.png", {
        maxZoom: 20,
        attribution:
          '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> &copy; <a href="https://carto.com/attributions">CARTO</a>',
        className: "map-tiles-mono",
      }).addTo(map);
      L.tileLayer("https://{s}.basemaps.cartocdn.com/light_only_labels/{z}/{x}/{y}{r}.png", {
        maxZoom: 20,
        className: "map-tiles-labels",
      }).addTo(map);
      L.control.zoom({ position: "bottomright" }).addTo(map);

      map.on("click", (e) => {
        propsRef.current.onMapClick({ lat: round(e.latlng.lat), lng: round(e.latlng.lng) });
      });

      mapRef.current = map;
      setTimeout(() => map.invalidateSize(), 150);

      propsRef.current.registerFit(() => {
        if (optimizedRef.current) {
          map.flyToBounds(optimizedRef.current.getBounds(), { padding: [80, 80], duration: 0.9 });
        }
      });

      drawMarkers();
    });

    return () => {
      cancelled = true;
      if (mapRef.current) {
        mapRef.current.remove();
        mapRef.current = null;
      }
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const drawMarkers = () => {
    const L = leafletRef.current;
    const map = mapRef.current;
    if (!L || !map) return;

    markersRef.current.forEach((m) => map.removeLayer(m));
    markersRef.current = [];

    const { start, stops, end } = propsRef.current.route;

    const icon = (role: string, label: string) =>
      L.divIcon({
        className: "",
        html: `<div class="marker-pin marker-pin--${role}"><span>${label}</span></div>`,
        iconSize: [28, 28],
        iconAnchor: [14, 28],
      });

    const add = (coord: Coordinate, role: string, label: string, onMove: (c: Coordinate) => void) => {
      const marker = L.marker([coord.lat, coord.lng], { icon: icon(role, label), draggable: true }).addTo(map);
      marker.on("dragend", () => {
        const p = marker.getLatLng();
        onMove({ lat: round(p.lat), lng: round(p.lng) });
      });
      markersRef.current.push(marker);
    };

    if (start) add(start, "start", "A", (c) => propsRef.current.onDragStart(c));
    stops.forEach((stop, i) => add(stop, "stop", String(i + 1), (c) => propsRef.current.onDragStop(i, c)));
    if (end) add(end, "end", "B", (c) => propsRef.current.onDragEnd(c));
  };

  useEffect(() => {
    drawMarkers();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [props.route]);

  useEffect(() => {
    const L = leafletRef.current;
    const map = mapRef.current;
    if (!L || !map) return;

    if (originalRef.current) {
      map.removeLayer(originalRef.current);
      originalRef.current = null;
    }
    if (optimizedRef.current) {
      map.removeLayer(optimizedRef.current);
      optimizedRef.current = null;
    }

    const toLatLng = (coords: number[][]) => coords.map((c) => [c[1], c[0]] as [number, number]);

    if (props.originalGeometry && props.originalGeometry.length && props.showOriginal) {
      originalRef.current = L.polyline(toLatLng(props.originalGeometry), {
        color: "#8b93a1",
        weight: 4,
        dashArray: "6 9",
        opacity: 0.85,
      }).addTo(map);
    }

    if (props.optimizedGeometry && props.optimizedGeometry.length && props.showOptimized) {
      optimizedRef.current = L.polyline(toLatLng(props.optimizedGeometry), {
        color: "#22c55e",
        weight: 6,
        opacity: 1,
        lineJoin: "round",
      }).addTo(map);
    }
  }, [props.optimizedGeometry, props.originalGeometry, props.showOptimized, props.showOriginal]);

  useEffect(() => {
    const L = leafletRef.current;
    const map = mapRef.current;
    if (!L || !map) return;

    hexRef.current.forEach((h) => map.removeLayer(h));
    hexRef.current = [];

    if (!props.showHexGrid) return;

    props.clusters.forEach((cluster) => {
      const ring = cluster.boundary.map((c) => [c[1], c[0]] as [number, number]);
      if (ring.length < 4) return;
      const polygon = L.polygon(ring, {
        color: "#16a34a",
        weight: 1.2,
        fillColor: "#22c55e",
        fillOpacity: Math.min(0.08 + cluster.count * 0.06, 0.32),
      }).addTo(map);
      hexRef.current.push(polygon);
    });
  }, [props.clusters, props.showHexGrid]);

  return <div ref={containerRef} className="map-root" role="application" aria-label="Delivery route map centered on Nairobi" />;
}
