import { useState, useEffect } from 'react';
import type { Filter, FilterFormData } from '../types';
import { X, ArrowLeftRight, Globe } from 'lucide-react';
import { AirportAutocomplete } from './AirportAutocomplete';

interface FilterFormProps {
  filter?: Filter | null;
  onSubmit: (data: Omit<Filter, 'id' | 'userId'>) => Promise<void>;
  onClose: () => void;
}

function toFormData(f?: Filter | null): FilterFormData {
  return {
    origin: f?.origin ?? '',
    destination: f?.destination ?? '',
    priceMax: f?.priceMax ?? 0,
    mode: (f?.mode as 'flight' | 'train') ?? 'flight',
    startDate: f?.startDate ? f.startDate.slice(0, 10) : '',
    endDate: f?.endDate ? f.endDate.slice(0, 10) : '',
    passengers: f?.passengers ?? 1,
    maxDurationHours: f?.maxDurationHours ?? 0,
    maxStops: f?.maxStops ?? 0,
    preferredDeparture: f?.preferredDeparture ?? '',
    preferredArrival: f?.preferredArrival ?? '',
    preferredAirlines: f?.preferredAirlines?.join(', ') ?? '',
    excludedAirlines: f?.excludedAirlines?.join(', ') ?? '',
    notifyEmail: f?.notifyEmail ?? '',
    notifyPushToken: f?.notifyPushToken ?? '',
    isActive: f?.isActive ?? true,
  };
}

export function FilterForm({ filter, onSubmit, onClose }: FilterFormProps) {
  const [form, setForm] = useState<FilterFormData>(() => toFormData(filter));
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    setForm(toFormData(filter));
  }, [filter]);

  const update = <K extends keyof FilterFormData>(key: K, value: FilterFormData[K]) => {
    setForm((prev) => ({ ...prev, [key]: value }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    try {
      await onSubmit({
        origin: form.origin,
        destination: form.destination,
        priceMax: form.priceMax,
        mode: form.mode,
        startDate: form.startDate,
        endDate: form.endDate,
        passengers: form.passengers,
        maxDurationHours: form.maxDurationHours || undefined,
        maxStops: form.maxStops || undefined,
        preferredDeparture: form.preferredDeparture || undefined,
        preferredArrival: form.preferredArrival || undefined,
        preferredAirlines: form.preferredAirlines ? form.preferredAirlines.split(',').map((s) => s.trim()).filter(Boolean) : undefined,
        excludedAirlines: form.excludedAirlines ? form.excludedAirlines.split(',').map((s) => s.trim()).filter(Boolean) : undefined,
        notifyEmail: form.notifyEmail || undefined,
        notifyPushToken: form.notifyPushToken || undefined,
        isActive: form.isActive,
      });
      onClose();
    } finally {
      setSaving(false);
    }
  };

  const label = 'block text-sm font-medium text-gray-700 mb-1';
  const input = 'w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-shadow';
  const half = 'grid grid-cols-1 sm:grid-cols-2 gap-4';

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center pt-10 sm:pt-20 bg-black/40 overflow-y-auto">
      <div className="relative bg-white rounded-xl shadow-xl w-full max-w-2xl mx-4 mb-10">
        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold text-gray-900">
            {filter ? 'Editar filtro' : 'Novo filtro'}
          </h2>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600 transition-colors">
            <X className="h-5 w-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-6 space-y-6">
          <fieldset>
            <legend className="text-sm font-semibold text-gray-900 mb-3">Rotas</legend>
            <div className={half}>
              <div>
                <label className={label} htmlFor="origin">Origem</label>
                <div className="flex gap-2">
                  <div className="flex-1">
                    <AirportAutocomplete
                      value={form.origin}
                      onChange={(iata) => update('origin', iata === 'ANY' ? 'ANY' : iata)}
                      placeholder="Cidade ou aeroporto"
                      otherValue={form.destination}
                    />
                  </div>
                  <button
                    type="button"
                    onClick={() => {
                      const orig = form.origin;
                      const dest = form.destination;
                      update('origin', dest);
                      update('destination', orig);
                    }}
                    className="self-end p-2 text-gray-400 hover:text-blue-600 hover:bg-blue-50 rounded-lg transition-colors"
                    title="Inverter origem/destino"
                  >
                    <ArrowLeftRight className="h-4 w-4" />
                  </button>
                </div>
                <label className="flex items-center gap-1.5 mt-1.5 text-xs text-gray-400 cursor-pointer hover:text-blue-600 transition-colors">
                  <input
                    type="checkbox"
                    checked={form.origin === 'ANY'}
                    onChange={(e) => update('origin', e.target.checked ? 'ANY' : '')}
                    className="h-3.5 w-3.5 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                  />
                  <Globe className="h-3 w-3" />
                  Qualquer lugar
                </label>
              </div>
              <div>
                <label className={label} htmlFor="destination">Destino</label>
                <AirportAutocomplete
                  value={form.destination}
                  onChange={(iata) => update('destination', iata === 'ANY' ? 'ANY' : iata)}
                  placeholder="Cidade ou aeroporto"
                  otherValue={form.origin}
                />
                <label className="flex items-center gap-1.5 mt-1.5 text-xs text-gray-400 cursor-pointer hover:text-blue-600 transition-colors">
                  <input
                    type="checkbox"
                    checked={form.destination === 'ANY'}
                    onChange={(e) => update('destination', e.target.checked ? 'ANY' : '')}
                    className="h-3.5 w-3.5 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                  />
                  <Globe className="h-3 w-3" />
                  Qualquer lugar
                </label>
              </div>
            </div>
          </fieldset>

          <fieldset>
            <legend className="text-sm font-semibold text-gray-900 mb-3">Preço e Datas</legend>
            <div className={half}>
              <div>
                <label className={label} htmlFor="priceMax">Preço máximo (R$)</label>
                <input id="priceMax" type="number" min={0} step={0.01} className={input} value={form.priceMax} onChange={(e) => update('priceMax', Number(e.target.value))} required />
              </div>
              <div>
                <label className={label} htmlFor="passengers">Passageiros</label>
                <input id="passengers" type="number" min={1} className={input} value={form.passengers} onChange={(e) => update('passengers', Number(e.target.value))} required />
              </div>
              <div>
                <label className={label} htmlFor="startDate">Data ida</label>
                <input id="startDate" type="date" className={input} value={form.startDate} onChange={(e) => update('startDate', e.target.value)} />
              </div>
              <div>
                <label className={label} htmlFor="endDate">Data volta</label>
                <input id="endDate" type="date" className={input} value={form.endDate} onChange={(e) => update('endDate', e.target.value)} />
              </div>
            </div>
          </fieldset>

          <fieldset>
            <legend className="text-sm font-semibold text-gray-900 mb-3">Preferências de voo</legend>
            <div className={half}>
              <div>
                <label className={label} htmlFor="maxStops">Máx. escalas</label>
                <input id="maxStops" type="number" min={0} className={input} value={form.maxStops} onChange={(e) => update('maxStops', Number(e.target.value))} />
              </div>
              <div>
                <label className={label} htmlFor="maxDurationHours">Duração máx. (horas)</label>
                <input id="maxDurationHours" type="number" min={0} className={input} value={form.maxDurationHours} onChange={(e) => update('maxDurationHours', Number(e.target.value))} />
              </div>
              <div>
                <label className={label} htmlFor="preferredDeparture">Partida preferida</label>
                <input id="preferredDeparture" type="time" className={input} value={form.preferredDeparture} onChange={(e) => update('preferredDeparture', e.target.value)} />
              </div>
              <div>
                <label className={label} htmlFor="preferredArrival">Chegada preferida</label>
                <input id="preferredArrival" type="time" className={input} value={form.preferredArrival} onChange={(e) => update('preferredArrival', e.target.value)} />
              </div>
              <div>
                <label className={label} htmlFor="preferredAirlines">Cias. preferidas (separadas por vírgula)</label>
                <input id="preferredAirlines" className={input} value={form.preferredAirlines} onChange={(e) => update('preferredAirlines', e.target.value)} placeholder="TAP, LATAM" />
              </div>
              <div>
                <label className={label} htmlFor="excludedAirlines">Cias. excluídas (separadas por vírgula)</label>
                <input id="excludedAirlines" className={input} value={form.excludedAirlines} onChange={(e) => update('excludedAirlines', e.target.value)} placeholder="GOL" />
              </div>
            </div>
          </fieldset>

          <fieldset>
            <legend className="text-sm font-semibold text-gray-900 mb-3">Notificações</legend>
            <div className={half}>
              <div>
                <label className={label} htmlFor="notifyEmail">E-mail para notificação</label>
                <input id="notifyEmail" type="email" className={input} value={form.notifyEmail} onChange={(e) => update('notifyEmail', e.target.value)} placeholder="email@exemplo.com" />
              </div>
              <div>
                <label className={label} htmlFor="notifyPushToken">Push token</label>
                <input id="notifyPushToken" className={input} value={form.notifyPushToken} onChange={(e) => update('notifyPushToken', e.target.value)} />
              </div>
            </div>
          </fieldset>

          <div className="flex items-center gap-2">
            <input
              id="isActive"
              type="checkbox"
              className="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
              checked={form.isActive}
              onChange={(e) => update('isActive', e.target.checked)}
            />
            <label htmlFor="isActive" className="text-sm text-gray-700">Filtro ativo</label>
          </div>

          <div className="flex items-center justify-end gap-3 pt-4 border-t border-gray-200">
            <button type="button" onClick={onClose} className="px-4 py-2 text-sm text-gray-700 hover:text-gray-900 transition-colors">
              Cancelar
            </button>
            <button
              type="submit"
              disabled={saving}
              className="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 disabled:opacity-50 transition-colors"
            >
              {saving ? 'Salvando...' : filter ? 'Salvar alterações' : 'Criar filtro'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
