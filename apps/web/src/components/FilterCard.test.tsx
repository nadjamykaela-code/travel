import { render, screen } from '@testing-library/react';
import { FilterCard } from './FilterCard';
import type { Filter } from '../types';
import { describe, it, expect, vi } from 'vitest';

const baseFilter: Filter = {
  id: 'f1',
  userId: 'u1',
  origin: 'GRU',
  destination: 'LIS',
  priceMax: 3000,
  passengers: 2,
  isActive: true,
  startDate: '2026-08-15',
  endDate: '2026-08-20',
  notifyEmail: 'test@example.com',
};

describe('FilterCard', () => {
  it('renders origin and destination', () => {
    render(<FilterCard filter={baseFilter} onEdit={vi.fn()} onDelete={vi.fn()} />);
    expect(screen.getByText('GRU')).toBeInTheDocument();
    expect(screen.getByText('LIS')).toBeInTheDocument();
  });

  it('renders price', () => {
    render(<FilterCard filter={baseFilter} onEdit={vi.fn()} onDelete={vi.fn()} />);
    expect(screen.getByText(/Até R\$ 3000/)).toBeInTheDocument();
  });

  it('renders passengers', () => {
    render(<FilterCard filter={baseFilter} onEdit={vi.fn()} onDelete={vi.fn()} />);
    expect(screen.getByText(/2 passageiro/)).toBeInTheDocument();
  });

  it('shows active badge', () => {
    render(<FilterCard filter={baseFilter} onEdit={vi.fn()} onDelete={vi.fn()} />);
    expect(screen.getByText('Ativo')).toBeInTheDocument();
  });

  it('shows inactive badge', () => {
    render(<FilterCard filter={{ ...baseFilter, isActive: false }} onEdit={vi.fn()} onDelete={vi.fn()} />);
    expect(screen.getByText('Inativo')).toBeInTheDocument();
  });

  it('calls onEdit when edit button clicked', () => {
    const onEdit = vi.fn();
    render(<FilterCard filter={baseFilter} onEdit={onEdit} onDelete={vi.fn()} />);
    screen.getByText('Editar').click();
    expect(onEdit).toHaveBeenCalledWith(baseFilter);
  });

  it('calls onDelete with id when delete clicked', () => {
    const onDelete = vi.fn();
    render(<FilterCard filter={baseFilter} onEdit={vi.fn()} onDelete={onDelete} />);
    screen.getByText('Excluir').click();
    expect(onDelete).toHaveBeenCalledWith('f1');
  });

  it('renders date when startDate provided', () => {
    render(<FilterCard filter={baseFilter} onEdit={vi.fn()} onDelete={vi.fn()} />);
    expect(screen.getByText(/15/)).toBeInTheDocument();
  });
});
