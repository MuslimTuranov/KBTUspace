import { useState } from 'react';
import { Link } from 'react-router-dom';
import { MapPin, Calendar, Users, Globe, Building2, Pin, Flag } from 'lucide-react';
import { format, isPast } from 'date-fns';
import { useAuth } from '../context/AuthContext';
import { useFacultyName } from '../hooks/useFaculties';
import type { Event } from '../types';
import ReportModal from './ReportModal';

export default function EventCard({ event }: { event: Event }) {
  const { user } = useAuth();
  const [showReport, setShowReport] = useState(false);
  const isOwner = user?.id === event.author_id;
  const facultyName = useFacultyName(event.faculty_id);
  const past = isPast(new Date(event.event_date));
  const full = event.current_count >= event.capacity;
  return (
    <>
      <div className={`card p-4 flex flex-col gap-3 ${event.is_pinned ? 'border-blue-300 bg-blue-50/30' : ''}`}>
        <div className="flex items-start justify-between gap-2">
          <div className="flex flex-wrap items-center gap-2">
            {event.is_pinned && <span className="badge bg-blue-100 text-blue-700"><Pin className="w-3 h-3 mr-1" />Pinned</span>}
            <span className={`badge ${event.scope === 'global' ? 'bg-purple-100 text-purple-700' : 'bg-gray-100 text-gray-600'}`}>
              {event.scope === 'global' ? <Globe className="w-3 h-3 mr-1" /> : <Building2 className="w-3 h-3 mr-1" />}
              {event.scope === 'global' ? 'Global' : (facultyName || 'Faculty')}
            </span>
            {past && <span className="badge bg-gray-100 text-gray-500">Ended</span>}
            {full && !past && <span className="badge bg-red-100 text-red-600">Full</span>}
            {event.status === 'pending' && <span className="badge bg-yellow-100 text-yellow-700">Pending</span>}
          </div>
          {!isOwner && <button onClick={() => setShowReport(true)} className="btn-ghost p-1.5 rounded-md text-gray-400"><Flag className="w-4 h-4" /></button>}
        </div>
        {event.image_url && <img src={event.image_url} alt="" className="rounded-lg w-full h-40 object-cover" onError={(e) => { (e.target as HTMLImageElement).style.display = 'none'; }} />}
        <div>
          <Link to={`/events/${event.id}`} className="text-base font-semibold text-gray-900 hover:text-blue-700 line-clamp-2">{event.title}</Link>
          <p className="mt-1 text-sm text-gray-600 line-clamp-2">{event.description}</p>
        </div>
        <div className="flex flex-wrap gap-3 text-xs text-gray-500">
          <span className="flex items-center gap-1"><Calendar className="w-3.5 h-3.5" />{format(new Date(event.event_date), 'MMM d, yyyy · HH:mm')}</span>
          <span className="flex items-center gap-1"><MapPin className="w-3.5 h-3.5" />{event.location}</span>
          <span className="flex items-center gap-1"><Users className="w-3.5 h-3.5" />{event.current_count}/{event.capacity}</span>
        </div>
        <div className="flex items-center justify-between">
          <span className="text-xs text-gray-400">{format(new Date(event.created_at), 'MMM d, yyyy')}</span>
          <Link to={`/events/${event.id}`} className="btn-primary text-xs px-3 py-1.5">View</Link>
        </div>
      </div>
      {showReport && <ReportModal targetType="event" targetId={event.id} onClose={() => setShowReport(false)} />}
    </>
  );
}
