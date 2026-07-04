"use client";

import { useEffect, useRef, useState } from "react";
import { searchPlaces } from "@/lib/api";
import type { Place } from "@/lib/types";

interface SearchBoxProps {
  onSelect: (place: Place) => void;
}

export default function SearchBox({ onSelect }: SearchBoxProps) {
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<Place[]>([]);
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const abortRef = useRef<AbortController | null>(null);

  useEffect(() => {
    if (query.trim().length < 3) {
      setResults([]);
      return;
    }

    const handle = setTimeout(async () => {
      abortRef.current?.abort();
      const controller = new AbortController();
      abortRef.current = controller;
      setLoading(true);
      try {
        const places = await searchPlaces(query, controller.signal);
        setResults(places);
        setOpen(true);
      } catch {
        setResults([]);
      } finally {
        setLoading(false);
      }
    }, 320);

    return () => clearTimeout(handle);
  }, [query]);

  const choose = (place: Place) => {
    onSelect(place);
    setQuery("");
    setResults([]);
    setOpen(false);
  };

  return (
    <div className="search">
      <div className="search__field">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true">
          <circle cx="11" cy="11" r="7" />
          <path d="M21 21l-4.3-4.3" />
        </svg>
        <input
          className="search__input"
          type="text"
          placeholder="Search a place in Nairobi…"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onFocus={() => results.length && setOpen(true)}
          aria-label="Search locations"
        />
        {loading && <span className="search__spinner" aria-hidden="true" />}
      </div>

      {open && results.length > 0 && (
        <ul className="search__results" role="listbox">
          {results.map((place, i) => (
            <li key={`${place.coordinate.lat}-${i}`}>
              <button type="button" className="search__result" onClick={() => choose(place)}>
                <span className="search__result-name">{place.name}</span>
                <span className="search__result-meta">{place.display_name}</span>
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
