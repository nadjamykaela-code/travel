import type { Filter } from '../types';
import { Pencil, Trash2, MapPin, Calendar, DollarSign, Users } from 'lucide-react';

interface FilterCardProps {
  filter: Filter;
  onEdit: (f: Filter) => void;
  onDelete: (id: string) => void;
}

export function FilterCard({ filter, onEdit, onDelete }: FilterCardProps) {
  return (
    <div className="bg-white rounded-lg border border-gray-200 p-4 hover:shadow-sm transition-shadow">
      <div className="flex items-start justify-between mb-3">
        <div className="flex items-center gap-1.5 text-lg font-semibold text-gray-900">
          <MapPin className="h-4 w-4 text-blue-500" />
          <span>{filter.origin}</span>
          <span className="text-gray-400">→</span>
          <MapPin className="h-4 w-4 text-green-500" />
          <span>{filter.destination}</span>
        </div>
        <div className="flex items-center gap-1">
          <span
            className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${
              filter.isActive
                ? 'bg-green-100 text-green-800'
                : 'bg-gray-100 text-gray-600'
            }`}
          >
            {filter.isActive ? 'Ativo' : 'Inativo'}
          </span>
        </div>
      </div>

      <div className="grid grid-cols-2 sm:grid-cols-3 gap-2 text-sm text-gray-600 mb-4">
        <div className="flex items-center gap-1.5">
          <DollarSign className="h-3.5 w-3.5 text-gray-400" />
          <span>Até R$ {filter.priceMax}</span>
        </div>
        <div className="flex items-center gap-1.5">
          <Users className="h-3.5 w-3.5 text-gray-400" />
          <span>{filter.passengers} passageiro(s)</span>
        </div>
        {filter.startDate && (
          <div className="flex items-center gap-1.5">
            <Calendar className="h-3.5 w-3.5 text-gray-400" />
            <span className="truncate">{new Date(filter.startDate).toLocaleDateString('pt-BR')}</span>
          </div>
        )}
      </div>

      <div className="flex items-center justify-end gap-2 pt-3 border-t border-gray-100">
        <button
          onClick={() => onEdit(filter)}
          className="inline-flex items-center gap-1 text-sm text-gray-500 hover:text-blue-600 transition-colors"
        >
          <Pencil className="h-3.5 w-3.5" />
          Editar
        </button>
        <button
          onClick={() => onDelete(filter.id!)}
          className="inline-flex items-center gap-1 text-sm text-gray-500 hover:text-red-600 transition-colors"
        >
          <Trash2 className="h-3.5 w-3.5" />
          Excluir
        </button>
      </div>
    </div>
  );
}
