import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { ArrowLeft, Globe, Building2, Pin, Pencil, Trash2, Flag, Loader2 } from 'lucide-react';
import { format } from 'date-fns';
import { getPost, deletePost, pinPost } from '../api/posts';
import { useAuth } from '../context/AuthContext';
import { useFacultyName } from '../hooks/useFaculties';
import { useState } from 'react';
import EditPostModal from '../components/EditPostModal';
import ReportModal from '../components/ReportModal';

export default function PostDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { user } = useAuth();
  const qc = useQueryClient();
  const [showEdit, setShowEdit] = useState(false);
  const [showReport, setShowReport] = useState(false);
  const { data: post, isLoading } = useQuery({ queryKey: ['post', Number(id)], queryFn: () => getPost(Number(id)), enabled: !!id });
  const facultyName = useFacultyName(post?.faculty_id);
  const deleteMut = useMutation({ mutationFn: () => deletePost(Number(id)), onSuccess: () => { qc.invalidateQueries({ queryKey: ['posts'] }); navigate('/'); } });
  const pinMut = useMutation({ mutationFn: () => pinPost(Number(id), !post?.is_pinned), onSuccess: () => qc.invalidateQueries({ queryKey: ['post', Number(id)] }) });
  if (isLoading) return <div className="flex items-center justify-center py-16"><Loader2 className="w-8 h-8 animate-spin text-blue-600" /></div>;
  if (!post) return <div className="text-center py-16 text-gray-400">Post not found</div>;
  const isAuthor = user?.id === post.author_id;
  const canPin = user?.role === 'organizer' || user?.role === 'admin';
  return (
    <div className="max-w-2xl mx-auto">
      <button onClick={() => navigate(-1)} className="btn-ghost mb-4"><ArrowLeft className="w-4 h-4" /> Back</button>
      <div className="card p-6">
        <div className="flex items-start justify-between gap-2 mb-4">
          <div className="flex flex-wrap gap-2">
            {post.is_pinned && <span className="badge bg-blue-100 text-blue-700"><Pin className="w-3 h-3 mr-1" />Pinned</span>}
            <span className={`badge ${post.scope === 'global' ? 'bg-purple-100 text-purple-700' : 'bg-gray-100 text-gray-600'}`}>
              {post.scope === 'global' ? <Globe className="w-3 h-3 mr-1" /> : <Building2 className="w-3 h-3 mr-1" />}
              {post.scope === 'global' ? 'Global' : (facultyName || 'Faculty')}
            </span>
            {post.status === 'pending' && <span className="badge bg-yellow-100 text-yellow-700">Pending review</span>}
          </div>
          <div className="flex items-center gap-1">
            {canPin && <button onClick={() => pinMut.mutate()} className={`btn-ghost p-2 ${post.is_pinned ? 'text-blue-600' : ''}`}><Pin className="w-4 h-4" /></button>}
            {isAuthor && (<><button onClick={() => setShowEdit(true)} className="btn-ghost p-2"><Pencil className="w-4 h-4" /></button><button onClick={() => { if (confirm('Delete this post?')) deleteMut.mutate(); }} className="btn-ghost p-2 text-red-500 hover:bg-red-50"><Trash2 className="w-4 h-4" /></button></>)}
            {!isAuthor && <button onClick={() => setShowReport(true)} className="btn-ghost p-2 text-gray-400"><Flag className="w-4 h-4" /></button>}
          </div>
        </div>
        <h1 className="text-2xl font-bold text-gray-900 mb-2">{post.title}</h1>
        <p className="text-sm text-gray-400 mb-4">{format(new Date(post.created_at), 'MMMM d, yyyy · HH:mm')}</p>
        {post.image_url && <img src={post.image_url} alt="" className="rounded-xl w-full mb-4 object-cover max-h-80" onError={(e) => { (e.target as HTMLImageElement).style.display = 'none'; }} />}
        <div className="text-gray-700 leading-relaxed whitespace-pre-wrap">{post.content}</div>
      </div>
      {showEdit && <EditPostModal post={post} onClose={() => setShowEdit(false)} />}
      {showReport && <ReportModal targetType="post" targetId={post.id} onClose={() => setShowReport(false)} />}
    </div>
  );
}
