import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { updatePost } from '../api/posts';
import { uploadImage } from '../api/uploads';
import { getApiErrorMessage } from '../api/errors';
import type { Post } from '../types';
import Modal from './Modal';

const schema = z.object({
  title: z.string().min(3, 'Title must be at least 3 characters').max(255, 'Title must be at most 255 characters'),
  content: z.string().min(10, 'Content must be at least 10 characters').max(5000, 'Content must be at most 5000 characters'),
  scope: z.enum(['faculty', 'global']),
});
type FormValues = z.infer<typeof schema>;

export default function EditPostModal({ post, onClose }: { post: Post; onClose: () => void }) {
  const qc = useQueryClient();
  const [image, setImage] = useState<File | null>(null);
  const { register, handleSubmit, formState: { errors } } = useForm<FormValues>({ resolver: zodResolver(schema), defaultValues: { title: post.title, content: post.content, scope: post.scope } });
  const mut = useMutation({
    mutationFn: async (v: FormValues) => {
      const imageUrl = image ? await uploadImage(image) : post.image_url ?? undefined;
      return updatePost(post.id, { ...v, image_url: imageUrl });
    },
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['posts'] }); qc.invalidateQueries({ queryKey: ['post', post.id] }); onClose(); },
  });
  return (
    <Modal title="Edit Post" onClose={onClose}>
      <form onSubmit={handleSubmit((v) => mut.mutate(v))} className="space-y-4">
        <div><label className="block text-sm font-medium text-gray-700 mb-1">Title</label><input {...register('title')} className="input" />{errors.title && <p className="text-xs text-red-500 mt-1">{errors.title.message}</p>}</div>
        <div><label className="block text-sm font-medium text-gray-700 mb-1">Content</label><textarea {...register('content')} rows={5} className="input resize-none" />{errors.content && <p className="text-xs text-red-500 mt-1">{errors.content.message}</p>}</div>
        <div><label className="block text-sm font-medium text-gray-700 mb-1">Replace image (optional)</label><input type="file" accept="image/png,image/jpeg,image/gif,image/webp" onChange={(e) => setImage(e.target.files?.[0] ?? null)} className="input" /></div>
        <div><label className="block text-sm font-medium text-gray-700 mb-1">Scope</label><select {...register('scope')} className="input"><option value="faculty">Faculty only</option><option value="global">Global</option></select></div>
        {mut.error && <p className="text-sm text-red-500">{getApiErrorMessage(mut.error, 'Failed to update')}</p>}
        <div className="flex gap-2 justify-end"><button type="button" onClick={onClose} className="btn-secondary">Cancel</button><button type="submit" disabled={mut.isPending} className="btn-primary">{mut.isPending ? 'Saving...' : 'Save Changes'}</button></div>
      </form>
    </Modal>
  );
}
