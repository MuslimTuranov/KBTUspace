import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Plus, Globe, Building2, Loader2 } from 'lucide-react';
import { getEvents } from '../api/events';
import { useAuth } from '../context/AuthContext';
import { useFaculties } from '../hooks/useFaculties';
import EventCard from '../components/EventCard';
import CreateEventModal from '../components/CreateEventModal';

type Filter = 'faculty' | 'global';

export default function EventsPage() {
  const { user } = useAuth();
  const { data: faculties } = useFaculties();
  const [filter, setFilter] = useState<Filter>('faculty');
  const [selectedFacultyId, setSelectedFacultyId] = useState<number | null>(user?.faculty_id ?? null);
  const [showCreate, setShowCreate] = useState(false);
  const canCreate = user?.role === 'organizer' || user?.role === 'admin';
  const { data: events, isLoading, error } = useQuery({ queryKey: ['events', filter, selectedFacultyId], queryFn: () => getEvents(filter === 'global' ? { global: true } : { faculty_id: selectedFacultyId ?? undefined }), enabled: filter === 'global' || selectedFacultyId !== null });
  return (
    <div className="max-w-4xl mx-auto">
      <div className="flex items-center justify-between mb-6"><h1 className="text-xl font-bold text-gray-900">Events</h1>{canCreate && <button onClick={() => setShowCreate(true)} className="btn-primary"><Plus className="w-4 h-4" /> New Event</button>}</div>
      <div className="flex flex-wrap items-center gap-2 mb-6">
        <button onClick={() => setFilter('faculty')} className={`btn ${filter === 'faculty' ? 'btn-primary' : 'btn-secondary'}`}><Building2 className="w-4 h-4" /> Faculty</button>
        <button onClick={() => setFilter('global')} className={`btn ${filter === 'global' ? 'btn-primary' : 'btn-secondary'}`}><Globe className="w-4 h-4" /> Global</button>
        {filter === 'faculty' && <select value={selectedFacultyId ?? ''} onChange={(e) => setSelectedFacultyId(e.target.value ? Number(e.target.value) : null)} className="input py-2 text-sm max-w-xs"><option value="">— Select faculty —</option>{faculties?.map((f) => <option key={f.id} value={f.id}>{f.name}</option>)}</select>}
      </div>
      {filter === 'faculty' && !selectedFacultyId && <div className="text-center py-16 text-gray-400"><Building2 className="w-10 h-10 mx-auto mb-2 text-gray-300" /><p className="font-medium">Select a faculty above to see its events</p></div>}
      {isLoading && <div className="flex items-center justify-center py-16"><Loader2 className="w-8 h-8 animate-spin text-blue-600" /></div>}
      {error && <div className="p-4 bg-red-50 border border-red-200 rounded-lg text-red-600 text-sm">Failed to load events.</div>}
      {!isLoading && events?.length === 0 && (filter === 'global' || selectedFacultyId) && <div className="text-center py-16 text-gray-400"><p className="text-lg font-medium mb-1">No events yet</p>{canCreate && <p className="text-sm">Create the first event!</p>}</div>}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">{events?.map((event) => <EventCard key={event.id} event={event} />)}</div>
      {showCreate && <CreateEventModal onClose={() => setShowCreate(false)} />}
    </div>
  );
}
