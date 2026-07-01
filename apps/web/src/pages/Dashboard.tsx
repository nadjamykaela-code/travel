import { useState, useMemo } from 'react';
import { useFilters } from '../hooks/useFilters';
import { FilterCard } from '../components/FilterCard';
import { FilterForm } from '../components/FilterForm';
import { ConfirmDialog } from '../components/ConfirmDialog';
import { Spinner } from '../components/Spinner';
import { Plus, Search, Frown } from 'lucide-react';
import type { Filter } from '../types';

export function Dashboard() {
  const { filters, loading, error, create, update, remove, refresh } = useFilters();
  const [showForm, setShowForm] = useState(false);
  const [editing, setEditing] = useState<Filter | null>(null);
  const [deleting, setDeleting] = useState<string | null>(null);
  const [search, setSearch] = useState('');

  const filtered = useMemo(() => {
    if (!search.trim()) return filters;
    const q = search.toLowerCase();
    return filters.filter(
      (f) =>
        f.origin.toLowerCase().includes(q) ||
        f.destination.toLowerCase().includes(q)
    );
  }, [filters, search]);

  const activeFilters = useMemo(() => filtered.filter((f) => f.isActive), [filtered]);

  const handleEdit = (f: Filter) => {
    setEditing(f);
    setShowForm(true);
  };

  const handleDelete = async () => {
    if (!deleting) return;
    try {
      await remove(deleting);
    } finally {
      setDeleting(null);
    }
  };

  const handleFormSubmit = async (data: Omit<Filter, 'id' | 'userId'>) => {
    if (editing) {
      await update(editing.id!, data);
    } else {
      await create(data);
    }
  };

  const openCreate = () => {
    setEditing(null);
    setShowForm(true);
  };

  const closeForm = () => {
    setShowForm(false);
    setEditing(null);
  };

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center py-20">
        <Frown className="h-12 w-12 text-gray-400 mb-4" />
        <p className="text-gray-600 mb-4">{error}</p>
        <button onClick={refresh} className="text-sm text-blue-600 hover:text-blue-700 underline">
          Tentar novamente
        </button>
      </div>
    );
  }

  return (
    <div>
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
        <div>
          <h2 className="text-xl font-semibold text-gray-900">Filtros de busca</h2>
          <p className="text-sm text-gray-500 mt-1">
            {activeFilters.length} filtro(s) ativo(s)
          </p>
        </div>
        <button
          onClick={openCreate}
          className="inline-flex items-center gap-2 px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 transition-colors"
        >
          <Plus className="h-4 w-4" />
          Novo filtro
        </button>
      </div>

      <div className="relative mb-6">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
        <input
          type="text"
          placeholder="Buscar por origem ou destino..."
          className="w-full rounded-lg border border-gray-300 pl-10 pr-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
      </div>

      {loading ? (
        <div className="flex justify-center py-20">
          <Spinner className="h-8 w-8" />
        </div>
      ) : filtered.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-20 text-gray-400">
          <Search className="h-12 w-12 mb-4" />
          <p className="text-gray-600">
            {search ? 'Nenhum filtro encontrado para esta busca' : 'Nenhum filtro configurado ainda'}
          </p>
          {!search && (
            <button onClick={openCreate} className="mt-4 text-sm text-blue-600 hover:text-blue-700 underline">
              Criar primeiro filtro
            </button>
          )}
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {filtered.map((f) => (
            <FilterCard key={f.id} filter={f} onEdit={handleEdit} onDelete={(id) => setDeleting(id)} />
          ))}
        </div>
      )}

      {showForm && (
        <FilterForm filter={editing} onSubmit={handleFormSubmit} onClose={closeForm} />
      )}

      {deleting && (
        <ConfirmDialog
          title="Excluir filtro"
          message="Tem certeza que deseja excluir este filtro? Esta ação não pode ser desfeita."
          confirmLabel="Excluir"
          onConfirm={handleDelete}
          onCancel={() => setDeleting(null)}
        />
      )}
    </div>
  );
}
