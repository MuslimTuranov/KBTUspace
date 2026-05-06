import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { updatePost } from '../api/posts';
import type { Post } from '../types';
import Modal from './Modal';

const schema = z.object({ title: z.string().min(3).max(255), content: z.string().min(10).max(5000), image_url: z.string().url().optional().or(z.literal('')), scope: z.enum(['faculty', 'global']) });
type FormValues = z.infer<typeof schema>;

export default function EditPostModal({ post, onClose }: { post: Post; onClose: () => void }) {
  const qc = useQueryClient();
  const { register, handleSubmit, formState: { errors } } = useForm<FormValues>({ resolver: zodResolver(schema), defaultValues: { title: post.title, content: post.content, image_url: post.image_url ?? '', scope: post.scope } });
  const mut = useMutation({ mutationFn: (v: FormValues) => updatePost(post.id, { ...v, image_url: v.image_url || undefined }), onSuccess: () => { qc.invalidateQueries({ queryKey: ['posts'] }); qc.invalidateQueries({ queryKey: ['post', post.id] }); onClose(); } });
  return (
    <Modal title="Edit Post" onClose={onClose}>
      <form onSubmit={handleSubmit((v) => mut.mutate(v))} className="space-y-4">
        <div><label className="block text-sm font-medium text-gray-700 mb-1">Title</label><input {...register('title')} className="input" />{errors.title && <p className="text-xs text-red-500 mt-1">{errors.title.message}</p>}</div>
        <div><label className="block text-sm font-medium text-gray-700 mb-1">Content</label><textarea {...register('content')} rows={5} className="input resize-none" />{errors.content && <p className="text-xs text-red-500 mt-1">{errors.content.message}</p>}</div>
        <div><label className="block text-sm font-medium text-gray-700 mb-1">Image URL (optional)</label><input {...register('image_url')} className="input" placeholder="https://..." />{errors.image_url && <p className="text-xs text-red-500 mt-1">{errors.image_url.message}</p>}</div>
        <div><label className="block text-sm font-medium text-gray-700 mb-1">Scope</label><select {...register('scope')} className="input"><option value="faculty">Faculty only</option><option value="global">Global</option></select></div>
        {mut.error && <p className="text-sm text-red-500">{(mut.error as any).response?.data?.error || 'Failed to update'}</p>}
        <div className="flex gap-2 justify-end"><button type="button" onClick={onClose} className="btn-secondary">Cancel</button><button type="submit" disabled={mut.isPending} className="btn-primary">{mut.isPending ? 'Saving...' : 'Save Changes'}</button></div>
      </form>
    </Modal>
  );
}
