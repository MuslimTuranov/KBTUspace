import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Plus, Globe, Building2, Loader2 } from 'lucide-react';
import { getPosts } from '../api/posts';
import { useAuth } from '../context/AuthContext';
import { useFaculties } from '../hooks/useFaculties';
import PostCard from '../components/PostCard';
import CreatePostModal from '../components/CreatePostModal';

type Filter = 'faculty' | 'global';

export default function FeedPage() {
  const { user } = useAuth();
  const { data: faculties } = useFaculties();
  const [filter, setFilter] = useState<Filter>('faculty');
  const [selectedFacultyId, setSelectedFacultyId] = useState<number | null>(user?.faculty_id ?? null);
  const [showCreate, setShowCreate] = useState(false);
  const { data: posts, isLoading, error } = useQuery({ queryKey: ['posts', filter, selectedFacultyId], queryFn: () => getPosts(filter === 'global' ? { global: true } : { faculty_id: selectedFacultyId ?? undefined }), enabled: filter === 'global' || selectedFacultyId !== null });
  return (
    <div className="max-w-4xl mx-auto">
      <div className="flex items-center justify-between mb-6"><h1 className="text-xl font-bold text-gray-900">Feed</h1><button onClick={() => setShowCreate(true)} className="btn-primary"><Plus className="w-4 h-4" /> New Post</button></div>
      <div className="flex flex-wrap items-center gap-2 mb-6">
        <button onClick={() => setFilter('faculty')} className={`btn ${filter === 'faculty' ? 'btn-primary' : 'btn-secondary'}`}><Building2 className="w-4 h-4" /> Faculty</button>
        <button onClick={() => setFilter('global')} className={`btn ${filter === 'global' ? 'btn-primary' : 'btn-secondary'}`}><Globe className="w-4 h-4" /> Global</button>
        {filter === 'faculty' && <select value={selectedFacultyId ?? ''} onChange={(e) => setSelectedFacultyId(e.target.value ? Number(e.target.value) : null)} className="input py-2 text-sm max-w-xs"><option value="">— Select faculty —</option>{faculties?.map((f) => <option key={f.id} value={f.id}>{f.name}</option>)}</select>}
      </div>
      {filter === 'faculty' && !selectedFacultyId && <div className="text-center py-16 text-gray-400"><Building2 className="w-10 h-10 mx-auto mb-2 text-gray-300" /><p className="font-medium">Select a faculty above to see its posts</p></div>}
      {isLoading && <div className="flex items-center justify-center py-16"><Loader2 className="w-8 h-8 animate-spin text-blue-600" /></div>}
      {error && <div className="p-4 bg-red-50 border border-red-200 rounded-lg text-red-600 text-sm">Failed to load posts.</div>}
      {!isLoading && posts?.length === 0 && (filter === 'global' || selectedFacultyId) && <div className="text-center py-16 text-gray-400"><p className="text-lg font-medium mb-1">No posts yet</p><p className="text-sm">Be the first to post!</p></div>}
      <div className="max-w-2xl space-y-4">{posts?.map((post) => <PostCard key={post.id} post={post} />)}</div>
      {showCreate && <CreatePostModal onClose={() => setShowCreate(false)} />}
    </div>
  );
}
