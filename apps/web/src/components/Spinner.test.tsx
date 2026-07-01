import { render, screen } from '@testing-library/react';
import { Spinner } from './Spinner';
import { describe, it, expect } from 'vitest';

describe('Spinner', () => {
  it('renders with default className', () => {
    const { container } = render(<Spinner />);
    const svg = container.querySelector('svg');
    expect(svg).toBeInTheDocument();
    expect(svg).toHaveClass('animate-spin');
    expect(svg).toHaveClass('h-5');
    expect(svg).toHaveClass('w-5');
  });

  it('renders with custom className', () => {
    const { container } = render(<Spinner className="h-8 w-8" />);
    const svg = container.querySelector('svg');
    expect(svg).toHaveClass('h-8');
    expect(svg).toHaveClass('w-8');
  });

  it('has accessible label', () => {
    render(<Spinner />);
    expect(screen.getByLabelText('Carregando')).toBeInTheDocument();
  });
});
