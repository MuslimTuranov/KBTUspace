import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { ArrowLeft, Calendar, MapPin, Users, Globe, Building2, Pin, Trash2, Flag, Loader2 } from 'lucide-react';
import { format, isPast } from 'date-fns';
import { getEvent, registerForEvent, cancelEventRegistration, deleteEvent } from '../api/events';
import { useAuth } from '../context/AuthContext';
import { useFacultyName } from '../hooks/useFaculties';
import { useState } from 'react';
import ReportModal from '../components/ReportModal';

export default function EventDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { user } = useAuth();
  const qc = useQueryClient();
  const [showReport, setShowReport] = useState(false);
  const [registered, setRegistered] = useState(false);
  const { data: event, isLoading } = useQuery({ queryKey: ['event', Number(id)], queryFn: () => getEvent(Number(id)), enabled: !!id });
  const facultyName = useFacultyName(event?.faculty_id);
  const registerMut = useMutation({ mutationFn: () => registerForEvent(Number(id)), onSuccess: () => { setRegistered(true); qc.invalidateQueries({ queryKey: ['event', Number(id)] }); } });
  const cancelMut = useMutation({ mutationFn: () => cancelEventRegistration(Number(id)), onSuccess: () => { setRegistered(false); qc.invalidateQueries({ queryKey: ['event', Number(id)] }); } });
  const deleteMut = useMutation({ mutationFn: () => deleteEvent(Number(id)), onSuccess: () => { qc.invalidateQueries({ queryKey: ['events'] }); navigate('/events'); } });
  if (isLoading) return <div className="flex items-center justify-center py-16"><Loader2 className="w-8 h-8 animate-spin text-blue-600" /></div>;
  if (!event) return <div className="text-center py-16 text-gray-400">Event not found</div>;
  const isOwner = user?.id === event.author_id;
  const canManage = isOwner || user?.role === 'admin';
  const past = isPast(new Date(event.event_date));
  const full = event.current_count >= event.capacity;
  const spotsLeft = event.capacity - event.current_count;
  return (
    <div className="max-w-2xl mx-auto">
      <button onClick={() => navigate(-1)} className="btn-ghost mb-4"><ArrowLeft className="w-4 h-4" /> Back</button>
      <div className="card overflow-hidden">
        {event.image_url && <img src={event.image_url} alt="" className="w-full h-56 object-cover" onError={(e) => { (e.target as HTMLImageElement).style.display = 'none'; }} />}
        <div className="p-6">
          <div className="flex items-start justify-between gap-2 mb-4">
            <div className="flex flex-wrap gap-2">
              {event.is_pinned && <span className="badge bg-blue-100 text-blue-700"><Pin className="w-3 h-3 mr-1" />Pinned</span>}
              <span className={`badge ${event.scope === 'global' ? 'bg-purple-100 text-purple-700' : 'bg-gray-100 text-gray-600'}`}>
                {event.scope === 'global' ? <Globe className="w-3 h-3 mr-1" /> : <Building2 className="w-3 h-3 mr-1" />}
                {event.scope === 'global' ? 'Global' : (facultyName || 'Faculty')}
              </span>
              {past && <span className="badge bg-gray-100 text-gray-500">Ended</span>}
              {full && !past && <span className="badge bg-red-100 text-red-600">Full</span>}
              {event.status === 'pending' && <span className="badge bg-yellow-100 text-yellow-700">Pending approval</span>}
            </div>
            <div className="flex gap-1">
              {canManage && <button onClick={() => { if (confirm('Delete this event?')) deleteMut.mutate(); }} className="btn-ghost p-2 text-red-500 hover:bg-red-50"><Trash2 className="w-4 h-4" /></button>}
              {!isOwner && <button onClick={() => setShowReport(true)} className="btn-ghost p-2 text-gray-400"><Flag className="w-4 h-4" /></button>}
            </div>
          </div>
          <h1 className="text-2xl font-bold text-gray-900 mb-4">{event.title}</h1>
          <div className="flex flex-col gap-2 mb-4 text-sm text-gray-600">
            <span className="flex items-center gap-2"><Calendar className="w-4 h-4 text-blue-500" />{format(new Date(event.event_date), 'EEEE, MMMM d, yyyy · HH:mm')}</span>
            <span className="flex items-center gap-2"><MapPin className="w-4 h-4 text-blue-500" />{event.location}</span>
            <span className="flex items-center gap-2"><Users className="w-4 h-4 text-blue-500" />{event.current_count} registered · {spotsLeft > 0 ? `${spotsLeft} spots left` : 'No spots left'}</span>
          </div>
          <div className="w-full bg-gray-100 rounded-full h-2 mb-6"><div className={`h-2 rounded-full ${full ? 'bg-red-500' : 'bg-blue-500'}`} style={{ width: `${Math.min((event.current_count / event.capacity) * 100, 100)}%` }} /></div>
          <p className="text-gray-700 leading-relaxed whitespace-pre-wrap mb-6">{event.description}</p>
          {!past && !isOwner && (
            <div className="flex gap-3">
              {!registered ? <button onClick={() => registerMut.mutate()} disabled={registerMut.isPending || full} className="btn-primary flex-1 justify-center py-2.5">{registerMut.isPending ? 'Registering...' : full ? 'Event Full' : 'Register'}</button>
              : <button onClick={() => cancelMut.mutate()} disabled={cancelMut.isPending} className="btn-secondary flex-1 justify-center py-2.5">{cancelMut.isPending ? 'Cancelling...' : 'Cancel Registration'}</button>}
            </div>
          )}
          {registerMut.error && <p className="mt-2 text-sm text-red-500">{(registerMut.error as any).response?.data?.error || 'Failed to register'}</p>}
        </div>
      </div>
      {showReport && <ReportModal targetType="event" targetId={event.id} onClose={() => setShowReport(false)} />}
    </div>
  );
}
