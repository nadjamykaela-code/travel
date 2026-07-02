import axios from 'axios';
import type { Filter, Place } from '../types';

const api = axios.create({
  baseURL: '/api',
  headers: { 'Content-Type': 'application/json' },
});

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

export const filterService = {
  list: () => api.get<Filter[]>('/filters'),
  create: (data: Omit<Filter, 'id' | 'userId'>) => api.post<Filter>('/filters', data),
  update: (id: string, data: Partial<Filter>) => api.put<Filter>(`/filters/${id}`, data),
  remove: (id: string) => api.delete(`/filters/${id}`),
};

export const placeService = {
  search: (q: string) => api.get<Place[]>('/places/search', { params: { q } }),
};

export const authService = {
  verify: () => api.get<{ userId: string }>('/auth/verify'),
};

export default api;
