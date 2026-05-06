import { useForm, useWatch } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { AlertTriangle } from 'lucide-react';
import { createPost } from '../api/posts';
import { useAuth } from '../context/AuthContext';
import Modal from './Modal';

const schema = z.object({
  title: z.string().min(3).max(255),
  content: z.string().min(10).max(5000),
  image_url: z.string().url().optional().or(z.literal('')),
  scope: z.enum(['faculty', 'global']),
});
type FormValues = z.infer<typeof schema>;

export default function CreatePostModal({ onClose }: { onClose: () => void }) {
  const { user } = useAuth();
  const qc = useQueryClient();

  const { register, handleSubmit, control, formState: { errors } } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { scope: 'faculty' },
  });

  const scope = useWatch({ control, name: 'scope' });
  const noFaculty = !user?.faculty_id && scope === 'faculty';
  const canGlobal = user?.role === 'organizer' || user?.role === 'admin';

  const mut = useMutation({
    mutationFn: (v: FormValues) =>
      createPost({ ...v, image_url: v.image_url || undefined, faculty_id: user?.faculty_id ?? undefined }),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['posts'] }); onClose(); },
  });

  return (
    <Modal title="Create Post" onClose={onClose}>
      <form onSubmit={handleSubmit((v) => mut.mutate(v))} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Title</label>
          <input {...register('title')} className="input" placeholder="Post title..." />
          {errors.title && <p className="text-xs text-red-500 mt-1">{errors.title.message}</p>}
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Content</label>
          <textarea {...register('content')} rows={5} className="input resize-none" placeholder="Write your post..." />
          {errors.content && <p className="text-xs text-red-500 mt-1">{errors.content.message}</p>}
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Image URL (optional)</label>
          <input {...register('image_url')} className="input" placeholder="https://..." />
          {errors.image_url && <p className="text-xs text-red-500 mt-1">{errors.image_url.message}</p>}
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Scope</label>
          <select {...register('scope')} className="input">
            <option value="faculty">Faculty only</option>
            {canGlobal && (
              <option value="global">Global (requires admin approval)</option>
            )}
          </select>
          {!canGlobal && (
            <p className="text-xs text-gray-400 mt-1">Only organizers and admins can post globally.</p>
          )}
        </div>

        {noFaculty && (
          <div className="flex gap-2 p-3 bg-amber-50 border border-amber-200 rounded-lg text-sm text-amber-800">
            <AlertTriangle className="w-4 h-4 shrink-0 mt-0.5" />
            <span>
              You need a Faculty ID on your profile to post to faculty.{' '}
              <Link to="/profile" onClick={onClose} className="font-medium underline underline-offset-2">
                Set it now
              </Link>
            </span>
          </div>
        )}

        {mut.error && (
          <p className="text-sm text-red-500">
            {(mut.error as any).response?.data?.error || 'Failed to create post'}
          </p>
        )}

        <div className="flex gap-2 justify-end">
          <button type="button" onClick={onClose} className="btn-secondary">Cancel</button>
          <button type="submit" disabled={mut.isPending || noFaculty} className="btn-primary">
            {mut.isPending ? 'Creating...' : 'Create Post'}
          </button>
        </div>
      </form>
    </Modal>
  );
}
