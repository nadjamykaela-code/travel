import { useState, useRef, useEffect, useCallback } from 'react';
import { placeService } from '../services/api';
import type { Place } from '../types';
import { Search, MapPin, Plane, Loader2 } from 'lucide-react';

interface AirportAutocompleteProps {
  value: string;
  onChange: (iata: string, place?: Place) => void;
  placeholder?: string;
  otherValue?: string;
}

export function AirportAutocomplete({ value, onChange, placeholder, otherValue }: AirportAutocompleteProps) {
  const labelFor = (v: string) => v === 'ANY' ? 'Qualquer lugar' : v;

  const [query, setQuery] = useState(() => labelFor(value));
  const [results, setResults] = useState<Place[]>([]);
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const listRef = useRef<HTMLUListElement>(null);
  const debounceRef = useRef<ReturnType<typeof setTimeout>>();
  const prevValueRef = useRef(value);
  const selectingRef = useRef(false);

  const search = useCallback(async (q: string) => {
    if (q.length < 2) {
      setResults([]);
      setOpen(false);
      setError(null);
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const res = await placeService.search(q);
      setResults(res.data);
      setOpen(true);
      if (res.data.length === 0) {
        setError('Nenhum resultado encontrado');
      }
    } catch {
      setResults([]);
      setError('Não foi possível buscar sugestões, tente novamente');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (prevValueRef.current !== value) {
      prevValueRef.current = value;
      if (selectingRef.current) {
        selectingRef.current = false;
        return;
      }
      setQuery(labelFor(value));
    }
  }, [value]);

  const handleInput = (val: string) => {
    setQuery(val);

    if (debounceRef.current) clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(() => search(val), 300);
  };

  const select = (place: Place) => {
    selectingRef.current = true;
    setQuery(`${place.iataCode} — ${place.cityName} (${place.name})`);
    onChange(place.iataCode, place);
    setOpen(false);
    setError(null);
  };

  useEffect(() => {
    const handleClick = (e: MouseEvent) => {
      const target = e.target as Node;
      const insideInput = inputRef.current?.contains(target);
      const insideList = listRef.current?.contains(target);
      if (insideInput || insideList) return;

      // usuário digitou algo mas não escolheu uma sugestão: descarta o texto livre
      // e volta a mostrar o valor realmente selecionado, para não divergir do que será salvo.
      setOpen(false);
      setQuery(labelFor(value));
      setError(null);
    };
    document.addEventListener('mousedown', handleClick);
    return () => document.removeEventListener('mousedown', handleClick);
  }, [value]);

  return (
    <div className="relative">
      <div className="relative">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
        <input
          ref={inputRef}
          className="w-full rounded-lg border border-gray-300 pl-9 pr-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-shadow"
          value={query}
          onChange={(e) => handleInput(e.target.value)}
          onFocus={() => results.length > 0 && setOpen(true)}
          placeholder={placeholder}
          autoComplete="off"
        />
        {loading && (
          <Loader2 className="absolute right-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400 animate-spin" />
        )}
      </div>

      {!loading && error && (
        <p className="mt-1 text-xs text-red-500">{error}</p>
      )}

      {open && results.length > 0 && (
        <ul
          ref={listRef}
          className="absolute z-50 mt-1 w-full bg-white border border-gray-200 rounded-lg shadow-lg max-h-60 overflow-auto"
        >
          {results.map((place) => (
            <li
              key={place.entityId}
              className={`flex items-center gap-3 px-3 py-2.5 cursor-pointer hover:bg-blue-50 transition-colors text-sm ${
                place.iataCode === otherValue ? 'opacity-40 pointer-events-none' : ''
              }`}
              onClick={() => !(place.iataCode === otherValue) && select(place)}
            >
              <Plane className="h-4 w-4 text-gray-400 shrink-0" />
              <div className="min-w-0">
                <span className="font-semibold text-gray-900">{place.iataCode}</span>
                <span className="text-gray-600 ml-1.5">{place.name}</span>
                <div className="text-xs text-gray-400">
                  <MapPin className="inline h-3 w-3 mr-0.5" />
                  {place.cityName}, {place.countryName}
                </div>
              </div>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
