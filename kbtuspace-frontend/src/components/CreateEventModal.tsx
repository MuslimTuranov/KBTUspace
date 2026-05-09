import { useForm, useWatch } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { AlertTriangle } from 'lucide-react';
import { createEvent } from '../api/events';
import { uploadImage } from '../api/uploads';
import { getApiErrorMessage } from '../api/errors';
import { useAuth } from '../context/AuthContext';
import { useFaculties } from '../hooks/useFaculties';
import Modal from './Modal';
import { useState } from 'react';

const schema = z.object({ title: z.string().min(3, 'Title must be at least 3 characters').max(255, 'Title must be at most 255 characters'), description: z.string().min(10, 'Description must be at least 10 characters').max(5000, 'Description must be at most 5000 characters'), event_date: z.string().min(1, 'Date is required').refine((value) => new Date(value).getTime() > Date.now(), 'Event date cannot be in the past'), location: z.string().min(3, 'Location must be at least 3 characters').max(255, 'Location must be at most 255 characters'), capacity: z.number().int().min(1).max(10000), scope: z.enum(['faculty', 'global']) });
type FormValues = z.infer<typeof schema>;

export default function CreateEventModal({ onClose }: { onClose: () => void }) {
  const { user } = useAuth();
  const qc = useQueryClient();
  const [image, setImage] = useState<File | null>(null);
  const { data: faculties } = useFaculties();
  const [adminFacultyId, setAdminFacultyId] = useState<number | null>(null);
  const { register, handleSubmit, control, formState: { errors } } = useForm<FormValues>({ resolver: zodResolver(schema), defaultValues: { scope: 'faculty', capacity: 50 } });
  const scope = useWatch({ control, name: 'scope' });
  const facultyID = user?.role === 'admin' ? adminFacultyId : user?.faculty_id;
  const noFaculty = !facultyID && scope === 'faculty';
  const minEventDate = new Date(Date.now() - new Date().getTimezoneOffset() * 60000).toISOString().slice(0, 16);
  const mut = useMutation({ mutationFn: async (v: FormValues) => { const imageUrl = image ? await uploadImage(image) : undefined; return createEvent({ ...v, event_date: new Date(v.event_date).toISOString(), image_url: imageUrl, faculty_id: facultyID ?? undefined }); }, onSuccess: () => { qc.invalidateQueries({ queryKey: ['events'] }); onClose(); } });
  return (
    <Modal title="Create Event" onClose={onClose} size="lg">
      <form onSubmit={handleSubmit((v) => mut.mutate(v))} className="space-y-4">
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div className="sm:col-span-2"><label className="block text-sm font-medium text-gray-700 mb-1">Title</label><input {...register('title')} className="input" placeholder="Event title..." />{errors.title && <p className="text-xs text-red-500 mt-1">{errors.title.message}</p>}</div>
          <div className="sm:col-span-2"><label className="block text-sm font-medium text-gray-700 mb-1">Description</label><textarea {...register('description')} rows={4} className="input resize-none" placeholder="Describe the event..." />{errors.description && <p className="text-xs text-red-500 mt-1">{errors.description.message}</p>}</div>
          <div><label className="block text-sm font-medium text-gray-700 mb-1">Date & Time</label><input {...register('event_date')} type="datetime-local" min={minEventDate} className="input" />{errors.event_date && <p className="text-xs text-red-500 mt-1">{errors.event_date.message}</p>}</div>
          <div><label className="block text-sm font-medium text-gray-700 mb-1">Capacity</label><input {...register('capacity', { valueAsNumber: true })} type="number" min={1} max={10000} className="input" />{errors.capacity && <p className="text-xs text-red-500 mt-1">{errors.capacity.message}</p>}</div>
          <div className="sm:col-span-2"><label className="block text-sm font-medium text-gray-700 mb-1">Location</label><input {...register('location')} className="input" placeholder="Room / building / online..." />{errors.location && <p className="text-xs text-red-500 mt-1">{errors.location.message}</p>}</div>
          <div className="sm:col-span-2"><label className="block text-sm font-medium text-gray-700 mb-1">Image (optional)</label><input type="file" accept="image/png,image/jpeg,image/gif,image/webp" onChange={(e) => setImage(e.target.files?.[0] ?? null)} className="input" /></div>
          <div><label className="block text-sm font-medium text-gray-700 mb-1">Scope</label><select {...register('scope')} className="input"><option value="faculty">Faculty only</option><option value="global">{user?.role === 'admin' ? 'Global' : 'Global (requires approval)'}</option></select></div>
          {user?.role === 'admin' && scope === 'faculty' && <div><label className="block text-sm font-medium text-gray-700 mb-1">Faculty</label><select value={adminFacultyId ?? ''} onChange={(e) => setAdminFacultyId(e.target.value ? Number(e.target.value) : null)} className="input"><option value="">Select faculty</option>{faculties?.map(f => <option key={f.id} value={f.id}>{f.name}</option>)}</select></div>}
        </div>
        {noFaculty && <div className="flex gap-2 p-3 bg-amber-50 border border-amber-200 rounded-lg text-sm text-amber-800"><AlertTriangle className="w-4 h-4 shrink-0 mt-0.5" /><span>{user?.role === 'admin' ? 'Select a faculty to create a faculty event.' : <>You need a Faculty ID to create a faculty event. <Link to="/profile" onClick={onClose} className="font-medium underline underline-offset-2">Set it now</Link></>}</span></div>}
        {mut.error && <p className="text-sm text-red-500">{getApiErrorMessage(mut.error, 'Failed to create event')}</p>}
        <div className="flex gap-2 justify-end"><button type="button" onClick={onClose} className="btn-secondary">Cancel</button><button type="submit" disabled={mut.isPending || noFaculty} className="btn-primary">{mut.isPending ? 'Creating...' : 'Create Event'}</button></div>
      </form>
    </Modal>
  );
}
