import { useForm, useWatch } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { AlertTriangle } from 'lucide-react';
import { createPost } from '../api/posts';
import { uploadImage } from '../api/uploads';
import { getApiErrorMessage } from '../api/errors';
import { useAuth } from '../context/AuthContext';
import { useFaculties } from '../hooks/useFaculties';
import Modal from './Modal';
import { useState } from 'react';

const schema = z.object({
  title: z.string().min(3, 'Title must be at least 3 characters').max(255, 'Title must be at most 255 characters'),
  content: z.string().min(10, 'Content must be at least 10 characters').max(5000, 'Content must be at most 5000 characters'),
  scope: z.enum(['faculty', 'global']),
});
type FormValues = z.infer<typeof schema>;

export default function CreatePostModal({ onClose }: { onClose: () => void }) {
  const { user } = useAuth();
  const qc = useQueryClient();
  const [image, setImage] = useState<File | null>(null);
  const { data: faculties } = useFaculties();
  const [adminFacultyId, setAdminFacultyId] = useState<number | null>(null);

  const { register, handleSubmit, control, formState: { errors } } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { scope: 'faculty' },
  });

  const scope = useWatch({ control, name: 'scope' });
  const facultyID = user?.role === 'admin' ? adminFacultyId : user?.faculty_id;
  const noFaculty = !facultyID && scope === 'faculty';

  const mut = useMutation({
    mutationFn: async (v: FormValues) => {
      const imageUrl = image ? await uploadImage(image) : undefined;
      return createPost({ ...v, image_url: imageUrl, faculty_id: facultyID ?? undefined });
    },
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

        {user?.role === 'admin' && scope === 'faculty' && (
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Faculty</label>
            <select value={adminFacultyId ?? ''} onChange={(e) => setAdminFacultyId(e.target.value ? Number(e.target.value) : null)} className="input">
              <option value="">Select faculty</option>
              {faculties?.map(f => <option key={f.id} value={f.id}>{f.name}</option>)}
            </select>
          </div>
        )}

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Image (optional)</label>
          <input type="file" accept="image/png,image/jpeg,image/gif,image/webp" onChange={(e) => setImage(e.target.files?.[0] ?? null)} className="input" />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Scope</label>
          <select {...register('scope')} className="input">
            <option value="faculty">Faculty only</option>
            <option value="global">{user?.role === 'admin' ? 'Global' : 'Global (requires admin approval)'}</option>
          </select>
          {scope === 'global' && user?.role !== 'admin' && (
            <p className="text-xs text-gray-400 mt-1">Global posts are reviewed by admins before publication.</p>
          )}
        </div>

        {noFaculty && (
          <div className="flex gap-2 p-3 bg-amber-50 border border-amber-200 rounded-lg text-sm text-amber-800">
            <AlertTriangle className="w-4 h-4 shrink-0 mt-0.5" />
            <span>
              {user?.role === 'admin' ? 'Select a faculty to create a faculty post.' : <>You need a Faculty ID on your profile to post to faculty. <Link to="/profile" onClick={onClose} className="font-medium underline underline-offset-2">Set it now</Link></>}
            </span>
          </div>
        )}

        {mut.error && (
          <p className="text-sm text-red-500">
            {getApiErrorMessage(mut.error, 'Failed to create post')}
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
