import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { CheckCircle, XCircle, Trash2, MessageSquare, Loader2, Shield, Users } from 'lucide-react';
import { format } from 'date-fns';
import { getPendingContent, approvePost, rejectPost, adminDeletePost, approveEvent, rejectEvent, adminDeleteEvent, getReports, closeReport, updateUser, listUsers } from '../api/admin';
import type { Post, Event, Report, User, Role } from '../types';
import Modal from '../components/Modal';

type Tab = 'moderation' | 'reports' | 'users';

function RejectModal({ onSubmit, onClose }: { onSubmit: (reason: string) => void; onClose: () => void }) {
  const [reason, setReason] = useState('');
  return (
    <Modal title="Reject Content" onClose={onClose} size="sm">
      <div className="space-y-4">
        <div><label className="block text-sm font-medium text-gray-700 mb-1">Reason</label><textarea value={reason} onChange={(e) => setReason(e.target.value)} rows={3} className="input resize-none" placeholder="Explain why this content is being rejected..." /></div>
        <div className="flex gap-2 justify-end"><button onClick={onClose} className="btn-secondary">Cancel</button><button onClick={() => reason.trim() && onSubmit(reason)} disabled={!reason.trim()} className="btn-danger">Reject</button></div>
      </div>
    </Modal>
  );
}

function CloseReportModal({ report, onClose }: { report: Report; onClose: () => void }) {
  const qc = useQueryClient();
  const [note, setNote] = useState('');
  const [status, setStatus] = useState<'closed' | 'rejected'>('closed');
  const mut = useMutation({ mutationFn: () => closeReport(report.id, status, note), onSuccess: () => { qc.invalidateQueries({ queryKey: ['admin-reports'] }); onClose(); } });
  return (
    <Modal title="Close Report" onClose={onClose} size="sm">
      <div className="space-y-4">
        <div className="p-3 bg-gray-50 rounded-lg text-sm text-gray-700"><p className="font-medium mb-1">Report reason:</p><p>{report.reason}</p></div>
        <div><label className="block text-sm font-medium text-gray-700 mb-1">Decision</label><select value={status} onChange={(e) => setStatus(e.target.value as 'closed' | 'rejected')} className="input"><option value="closed">Close (content violates rules)</option><option value="rejected">Reject (report is invalid)</option></select></div>
        <div><label className="block text-sm font-medium text-gray-700 mb-1">Review note</label><textarea value={note} onChange={(e) => setNote(e.target.value)} rows={3} className="input resize-none" placeholder="Add a note..." /></div>
        <div className="flex gap-2 justify-end"><button onClick={onClose} className="btn-secondary">Cancel</button><button onClick={() => mut.mutate()} disabled={mut.isPending} className="btn-primary">{mut.isPending ? 'Saving...' : 'Submit'}</button></div>
      </div>
    </Modal>
  );
}

function ContentModerationTab() {
  const qc = useQueryClient();
  const [rejectTarget, setRejectTarget] = useState<{ type: 'post' | 'event'; id: number } | null>(null);
  const { data, isLoading } = useQuery({ queryKey: ['admin-pending'], queryFn: () => getPendingContent('all') });
  const approveMut = useMutation({ mutationFn: ({ type, id }: { type: 'post' | 'event'; id: number }) => type === 'post' ? approvePost(id) : approveEvent(id), onSuccess: () => qc.invalidateQueries({ queryKey: ['admin-pending'] }) });
  const handleReject = async (reason: string) => { if (!rejectTarget) return; if (rejectTarget.type === 'post') await rejectPost(rejectTarget.id, reason); else await rejectEvent(rejectTarget.id, reason); qc.invalidateQueries({ queryKey: ['admin-pending'] }); setRejectTarget(null); };
  const deleteMut = useMutation({ mutationFn: ({ type, id }: { type: 'post' | 'event'; id: number }) => type === 'post' ? adminDeletePost(id) : adminDeleteEvent(id), onSuccess: () => qc.invalidateQueries({ queryKey: ['admin-pending'] }) });
  if (isLoading) return <div className="flex justify-center py-8"><Loader2 className="w-6 h-6 animate-spin text-blue-600" /></div>;
  const total = (data?.posts?.length ?? 0) + (data?.events?.length ?? 0);
  if (total === 0) return <div className="text-center py-12 text-gray-400"><CheckCircle className="w-10 h-10 mx-auto mb-2 text-green-400" /><p className="font-medium">All clear — no pending content</p></div>;
  const renderItem = (item: Post | Event, type: 'post' | 'event') => (
    <div key={item.id} className="card p-4">
      <div className="flex items-start justify-between gap-3">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 mb-1"><span className="badge bg-yellow-100 text-yellow-700 capitalize">{type}</span><span className="badge bg-gray-100 text-gray-600">{item.scope}</span></div>
          <p className="font-medium text-gray-900 truncate">{item.title}</p>
          <p className="text-sm text-gray-500 mt-1 line-clamp-2">{'content' in item ? item.content : item.description}</p>
          <p className="text-xs text-gray-400 mt-1">{format(new Date(item.created_at), 'MMM d, yyyy · HH:mm')}</p>
        </div>
        <div className="flex items-center gap-1 shrink-0">
          <button onClick={() => approveMut.mutate({ type, id: item.id })} disabled={approveMut.isPending} className="btn-ghost p-2 text-green-600 hover:bg-green-50"><CheckCircle className="w-5 h-5" /></button>
          <button onClick={() => setRejectTarget({ type, id: item.id })} className="btn-ghost p-2 text-red-500 hover:bg-red-50"><XCircle className="w-5 h-5" /></button>
          <button onClick={() => { if (confirm('Delete permanently?')) deleteMut.mutate({ type, id: item.id }); }} className="btn-ghost p-2 text-gray-400"><Trash2 className="w-5 h-5" /></button>
        </div>
      </div>
    </div>
  );
  return (
    <div className="space-y-6">
      {(data?.posts?.length ?? 0) > 0 && <div><h3 className="text-sm font-semibold text-gray-500 uppercase tracking-wide mb-3">Posts ({data?.posts.length})</h3><div className="space-y-3">{data?.posts.map((p) => renderItem(p, 'post'))}</div></div>}
      {(data?.events?.length ?? 0) > 0 && <div><h3 className="text-sm font-semibold text-gray-500 uppercase tracking-wide mb-3">Events ({data?.events.length})</h3><div className="space-y-3">{data?.events.map((e) => renderItem(e, 'event'))}</div></div>}
      {rejectTarget && <RejectModal onSubmit={handleReject} onClose={() => setRejectTarget(null)} />}
    </div>
  );
}

function ReportsTab() {
  const [selected, setSelected] = useState<Report | null>(null);
  const { data: reports, isLoading } = useQuery({ queryKey: ['admin-reports'], queryFn: () => getReports('pending') });
  if (isLoading) return <div className="flex justify-center py-8"><Loader2 className="w-6 h-6 animate-spin text-blue-600" /></div>;
  if (!reports || reports.length === 0) return <div className="text-center py-12 text-gray-400"><MessageSquare className="w-10 h-10 mx-auto mb-2 text-gray-300" /><p className="font-medium">No pending reports</p></div>;
  return (
    <div className="space-y-3">
      {reports.map((report) => (
        <div key={report.id} className="card p-4">
          <div className="flex items-start justify-between gap-3">
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2 mb-1"><span className="badge bg-orange-100 text-orange-700 capitalize">{report.target_type}</span><span className="badge bg-yellow-100 text-yellow-700">pending</span></div>
              <p className="font-medium text-gray-900 truncate">Re: "{report.target_title}"</p>
              <p className="text-sm text-gray-600 mt-1 line-clamp-2">{report.reason}</p>
              <p className="text-xs text-gray-400 mt-1">{format(new Date(report.created_at), 'MMM d, yyyy · HH:mm')}</p>
            </div>
            <button onClick={() => setSelected(report)} className="btn-secondary shrink-0">Review</button>
          </div>
        </div>
      ))}
      {selected && <CloseReportModal report={selected} onClose={() => setSelected(null)} />}
    </div>
  );
}

function UsersTab() {
  const qc = useQueryClient();
  const { data: users, isLoading } = useQuery({ queryKey: ['admin-users'], queryFn: listUsers });
  const [search, setSearch] = useState('');
  const mut = useMutation({ mutationFn: ({ id, data }: { id: number; data: Partial<Pick<User, 'role' | 'faculty_id' | 'is_banned'>> }) => updateUser(id, data), onSuccess: () => qc.invalidateQueries({ queryKey: ['admin-users'] }) });
  const roleBadge: Record<string, string> = { student: 'bg-green-100 text-green-700', organizer: 'bg-blue-100 text-blue-700', admin: 'bg-purple-100 text-purple-700' };
  const filtered = users?.filter(u => u.email.toLowerCase().includes(search.toLowerCase())) ?? [];
  if (isLoading) return <div className="flex justify-center py-8"><Loader2 className="w-6 h-6 animate-spin text-blue-600" /></div>;
  return (
    <div className="space-y-4">
      <input value={search} onChange={(e) => setSearch(e.target.value)} className="input" placeholder="Filter by email..." />
      {filtered.length === 0 && <div className="text-center py-12 text-gray-400"><Users className="w-10 h-10 mx-auto mb-2 text-gray-300" /><p className="font-medium">No users found</p></div>}
      <div className="space-y-2">
        {filtered.map((u) => (
          <div key={u.id} className="card p-4 flex items-center gap-3 flex-wrap">
            <div className="flex-1 min-w-0">
              <p className="font-medium text-gray-900 truncate">{u.email}</p>
              <div className="flex items-center gap-2 mt-1">
                <span className={`badge ${roleBadge[u.role] ?? 'bg-gray-100 text-gray-600'}`}>{u.role}</span>
                {u.faculty_id && <span className="badge bg-gray-100 text-gray-500">Faculty #{u.faculty_id}</span>}
                {u.is_banned && <span className="badge bg-red-100 text-red-600">Banned</span>}
              </div>
            </div>
            <div className="flex items-center gap-2 flex-wrap">
              <select value={u.role} onChange={(e) => mut.mutate({ id: u.id, data: { role: e.target.value as Role } })} className="input py-1 text-sm w-32"><option value="student">student</option><option value="organizer">organizer</option><option value="admin">admin</option></select>
              <button onClick={() => mut.mutate({ id: u.id, data: { is_banned: !u.is_banned } })} className={u.is_banned ? 'btn-secondary text-sm py-1' : 'btn-danger text-sm py-1'}>{u.is_banned ? 'Unban' : 'Ban'}</button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

export default function AdminPage() {
  const [tab, setTab] = useState<Tab>('moderation');
  return (
    <div className="max-w-3xl mx-auto">
      <div className="flex items-center gap-3 mb-6"><Shield className="w-6 h-6 text-purple-600" /><h1 className="text-xl font-bold text-gray-900">Admin Panel</h1></div>
      <div className="flex gap-2 mb-6 border-b border-gray-200">
        {(['moderation', 'reports', 'users'] as Tab[]).map((t) => (
          <button key={t} onClick={() => setTab(t)} className={`pb-3 px-1 text-sm font-medium border-b-2 transition-colors capitalize ${tab === t ? 'border-blue-600 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-700'}`}>
            {t === 'moderation' ? 'Content Moderation' : t === 'reports' ? 'Reports' : 'Users'}
          </button>
        ))}
      </div>
      {tab === 'moderation' && <ContentModerationTab />}
      {tab === 'reports' && <ReportsTab />}
      {tab === 'users' && <UsersTab />}
    </div>
  );
}
