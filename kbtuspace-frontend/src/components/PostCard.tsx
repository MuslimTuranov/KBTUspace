import { useState } from 'react';
import { Link } from 'react-router-dom';
import { Pin, Globe, Building2, Flag, Trash2, Pencil } from 'lucide-react';
import { formatDistanceToNow } from 'date-fns';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { deletePost, pinPost } from '../api/posts';
import { useAuth } from '../context/AuthContext';
import { useFacultyName } from '../hooks/useFaculties';
import type { Post } from '../types';
import ReportModal from './ReportModal';
import EditPostModal from './EditPostModal';

export default function PostCard({ post }: { post: Post }) {
  const { user } = useAuth();
  const qc = useQueryClient();
  const [showReport, setShowReport] = useState(false);
  const [showEdit, setShowEdit] = useState(false);
  const isAuthor = user?.id === post.author_id;
  const canPin = user?.role === 'organizer' || user?.role === 'admin';
  const facultyName = useFacultyName(post.faculty_id);
  const deleteMut = useMutation({ mutationFn: () => deletePost(post.id), onSuccess: () => qc.invalidateQueries({ queryKey: ['posts'] }) });
  const pinMut = useMutation({ mutationFn: () => pinPost(post.id, !post.is_pinned), onSuccess: () => qc.invalidateQueries({ queryKey: ['posts'] }) });
  return (
    <>
      <div className={`card p-4 flex flex-col gap-3 ${post.is_pinned ? 'border-blue-300 bg-blue-50/30' : ''}`}>
        <div className="flex items-start justify-between gap-2">
          <div className="flex flex-wrap items-center gap-2">
            {post.is_pinned && <span className="badge bg-blue-100 text-blue-700"><Pin className="w-3 h-3 mr-1" />Pinned</span>}
            <span className={`badge ${post.scope === 'global' ? 'bg-purple-100 text-purple-700' : 'bg-gray-100 text-gray-600'}`}>
              {post.scope === 'global' ? <Globe className="w-3 h-3 mr-1" /> : <Building2 className="w-3 h-3 mr-1" />}
              {post.scope === 'global' ? 'Global' : (facultyName || 'Faculty')}
            </span>
            {post.status === 'pending' && <span className="badge bg-yellow-100 text-yellow-700">Pending review</span>}
          </div>
          <div className="flex items-center gap-1 shrink-0">
            {canPin && <button onClick={() => pinMut.mutate()} disabled={pinMut.isPending} className={`btn-ghost p-1.5 rounded-md ${post.is_pinned ? 'text-blue-600' : ''}`} title={post.is_pinned ? 'Unpin' : 'Pin'}><Pin className="w-4 h-4" /></button>}
            {isAuthor && (<><button onClick={() => setShowEdit(true)} className="btn-ghost p-1.5 rounded-md"><Pencil className="w-4 h-4" /></button><button onClick={() => { if (confirm('Delete this post?')) deleteMut.mutate(); }} className="btn-ghost p-1.5 rounded-md text-red-500 hover:bg-red-50"><Trash2 className="w-4 h-4" /></button></>)}
            {!isAuthor && <button onClick={() => setShowReport(true)} className="btn-ghost p-1.5 rounded-md text-gray-400"><Flag className="w-4 h-4" /></button>}
          </div>
        </div>
        <div>
          <Link to={`/posts/${post.id}`} className="text-base font-semibold text-gray-900 hover:text-blue-700 line-clamp-2">{post.title}</Link>
          {post.image_url && <img src={post.image_url} alt="" className="mt-2 rounded-lg w-full h-48 object-cover" onError={(e) => { (e.target as HTMLImageElement).style.display = 'none'; }} />}
          <p className="mt-1 text-sm text-gray-600 line-clamp-3">{post.content}</p>
        </div>
        <div className="text-xs text-gray-400">{formatDistanceToNow(new Date(post.created_at), { addSuffix: true })}</div>
      </div>
      {showReport && <ReportModal targetType="post" targetId={post.id} onClose={() => setShowReport(false)} />}
      {showEdit && <EditPostModal post={post} onClose={() => setShowEdit(false)} />}
    </>
  );
}
