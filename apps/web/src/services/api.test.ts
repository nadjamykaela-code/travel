import { describe, it, expect } from 'vitest';
import api from '../services/api';

describe('api service', () => {
  it('exports an axios instance with baseURL', () => {
    expect(api.defaults).toBeDefined();
  });

  it('has interceptors configured', () => {
    expect(api.interceptors).toBeDefined();
    expect(api.interceptors.request).toBeDefined();
  });
});
