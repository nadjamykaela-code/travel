export interface Filter {
  id?: string;
  userId?: string;
  origin: string;
  destination: string;
  priceMax: number;
  mode?: 'flight' | 'train';
  startDate: string;
  endDate: string;
  passengers: number;
  maxDurationHours?: number;
  maxStops?: number;
  preferredDeparture?: string;
  preferredArrival?: string;
  preferredAirlines?: string[];
  excludedAirlines?: string[];
  notifyEmail?: string;
  notifyPushToken?: string;
  isActive: boolean;
}

export interface User {
  uid: string;
  email: string | null;
}

export interface AuthState {
  user: User | null;
  token: string | null;
  loading: boolean;
}

export interface FilterFormData {
  origin: string;
  destination: string;
  priceMax: number;
  mode: 'flight' | 'train';
  startDate: string;
  endDate: string;
  passengers: number;
  maxDurationHours: number;
  maxStops: number;
  preferredDeparture: string;
  preferredArrival: string;
  preferredAirlines: string;
  excludedAirlines: string;
  notifyEmail: string;
  notifyPushToken: string;
  isActive: boolean;
}
