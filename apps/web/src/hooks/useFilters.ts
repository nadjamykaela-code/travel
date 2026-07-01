import { useState, useEffect, useCallback } from 'react';
import type { Filter } from '../types';
import { filterService } from '../services/api';

interface UseFiltersReturn {
  filters: Filter[];
  loading: boolean;
  error: string | null;
  create: (data: Omit<Filter, 'id' | 'userId'>) => Promise<void>;
  update: (id: string, data: Partial<Filter>) => Promise<void>;
  remove: (id: string) => Promise<void>;
  refresh: () => Promise<void>;
}

export function useFilters(): UseFiltersReturn {
  const [filters, setFilters] = useState<Filter[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await filterService.list();
      setFilters(res.data);
    } catch {
      setError('Erro ao carregar filtros');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    refresh();
  }, [refresh]);

  const create = useCallback(async (data: Omit<Filter, 'id' | 'userId'>) => {
    const res = await filterService.create(data);
    setFilters((prev) => [...prev, res.data]);
  }, []);

  const update = useCallback(async (id: string, data: Partial<Filter>) => {
    const res = await filterService.update(id, data);
    setFilters((prev) => prev.map((f) => (f.id === id ? res.data : f)));
  }, []);

  const remove = useCallback(async (id: string) => {
    await filterService.remove(id);
    setFilters((prev) => prev.filter((f) => f.id !== id));
  }, []);

  return { filters, loading, error, create, update, remove, refresh };
}
