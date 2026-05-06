import { useForm, useWatch } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { AlertTriangle } from 'lucide-react';
import { createEvent } from '../api/events';
import { useAuth } from '../context/AuthContext';
import Modal from './Modal';

const schema = z.object({ title: z.string().min(3).max(255), description: z.string().min(10).max(5000), event_date: z.string().min(1, 'Date is required'), location: z.string().min(2).max(255), capacity: z.number().int().min(1).max(10000), image_url: z.string().url().optional().or(z.literal('')), scope: z.enum(['faculty', 'global']) });
type FormValues = z.infer<typeof schema>;

export default function CreateEventModal({ onClose }: { onClose: () => void }) {
  const { user } = useAuth();
  const qc = useQueryClient();
  const { register, handleSubmit, control, formState: { errors } } = useForm<FormValues>({ resolver: zodResolver(schema), defaultValues: { scope: 'faculty', capacity: 50 } });
  const scope = useWatch({ control, name: 'scope' });
  const noFaculty = !user?.faculty_id && scope === 'faculty';
  const mut = useMutation({ mutationFn: (v: FormValues) => createEvent({ ...v, event_date: new Date(v.event_date).toISOString(), image_url: v.image_url || undefined, faculty_id: user?.faculty_id ?? undefined }), onSuccess: () => { qc.invalidateQueries({ queryKey: ['events'] }); onClose(); } });
  return (
    <Modal title="Create Event" onClose={onClose} size="lg">
      <form onSubmit={handleSubmit((v) => mut.mutate(v))} className="space-y-4">
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div className="sm:col-span-2"><label className="block text-sm font-medium text-gray-700 mb-1">Title</label><input {...register('title')} className="input" placeholder="Event title..." />{errors.title && <p className="text-xs text-red-500 mt-1">{errors.title.message}</p>}</div>
          <div className="sm:col-span-2"><label className="block text-sm font-medium text-gray-700 mb-1">Description</label><textarea {...register('description')} rows={4} className="input resize-none" placeholder="Describe the event..." />{errors.description && <p className="text-xs text-red-500 mt-1">{errors.description.message}</p>}</div>
          <div><label className="block text-sm font-medium text-gray-700 mb-1">Date & Time</label><input {...register('event_date')} type="datetime-local" className="input" />{errors.event_date && <p className="text-xs text-red-500 mt-1">{errors.event_date.message}</p>}</div>
          <div><label className="block text-sm font-medium text-gray-700 mb-1">Capacity</label><input {...register('capacity', { valueAsNumber: true })} type="number" min={1} max={10000} className="input" />{errors.capacity && <p className="text-xs text-red-500 mt-1">{errors.capacity.message}</p>}</div>
          <div className="sm:col-span-2"><label className="block text-sm font-medium text-gray-700 mb-1">Location</label><input {...register('location')} className="input" placeholder="Room / building / online..." />{errors.location && <p className="text-xs text-red-500 mt-1">{errors.location.message}</p>}</div>
          <div className="sm:col-span-2"><label className="block text-sm font-medium text-gray-700 mb-1">Image URL (optional)</label><input {...register('image_url')} className="input" placeholder="https://..." />{errors.image_url && <p className="text-xs text-red-500 mt-1">{errors.image_url.message}</p>}</div>
          <div><label className="block text-sm font-medium text-gray-700 mb-1">Scope</label><select {...register('scope')} className="input"><option value="faculty">Faculty only</option><option value="global">Global (requires approval)</option></select></div>
        </div>
        {noFaculty && <div className="flex gap-2 p-3 bg-amber-50 border border-amber-200 rounded-lg text-sm text-amber-800"><AlertTriangle className="w-4 h-4 shrink-0 mt-0.5" /><span>You need a Faculty ID to create a faculty event. <Link to="/profile" onClick={onClose} className="font-medium underline underline-offset-2">Set it now</Link></span></div>}
        {mut.error && <p className="text-sm text-red-500">{(mut.error as any).response?.data?.error || 'Failed to create event'}</p>}
        <div className="flex gap-2 justify-end"><button type="button" onClick={onClose} className="btn-secondary">Cancel</button><button type="submit" disabled={mut.isPending || noFaculty} className="btn-primary">{mut.isPending ? 'Creating...' : 'Create Event'}</button></div>
      </form>
    </Modal>
  );
}
