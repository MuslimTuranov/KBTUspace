import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { format } from 'date-fns';
import { User, Mail, Shield, Building2 } from 'lucide-react';
import { updateProfile } from '../api/auth';
import { useAuth } from '../context/AuthContext';
import { useFaculties } from '../hooks/useFaculties';
import { useState } from 'react';

const schema = z.object({ email: z.string().email('Invalid email'), faculty_id: z.number().int().min(1).optional() });
type FormValues = z.infer<typeof schema>;
const roleBadgeClass: Record<string, string> = { student: 'bg-green-100 text-green-700', organizer: 'bg-blue-100 text-blue-700', admin: 'bg-purple-100 text-purple-700' };

export default function ProfilePage() {
  const { user, login } = useAuth();
  const { data: faculties } = useFaculties();
  const qc = useQueryClient();
  const [saved, setSaved] = useState(false);
  const { register, handleSubmit, formState: { errors, isSubmitting } } = useForm<FormValues>({ resolver: zodResolver(schema), defaultValues: { email: user?.email ?? '', faculty_id: user?.faculty_id ?? undefined } });
  const mut = useMutation({
    mutationFn: (v: FormValues) => updateProfile({ email: v.email, faculty_id: v.faculty_id ?? undefined }),
    onSuccess: async (_updatedUser) => {
      const token = localStorage.getItem('token');
      if (token) await login(token);
      qc.invalidateQueries({ queryKey: ['faculties'] });
      setSaved(true); setTimeout(() => setSaved(false), 3000);
    },
  });
  if (!user) return null;
  return (
    <div className="max-w-lg mx-auto">
      <h1 className="text-xl font-bold text-gray-900 mb-6">Profile</h1>
      <div className="card p-6 mb-6">
        <div className="flex items-center gap-4 mb-6">
          <div className="w-16 h-16 rounded-full bg-blue-100 flex items-center justify-center"><User className="w-8 h-8 text-blue-600" /></div>
          <div><p className="font-semibold text-gray-900">{user.email}</p>
            <div className="flex items-center gap-2 mt-1">
              <span className={`badge ${roleBadgeClass[user.role] ?? 'bg-gray-100 text-gray-600'}`}><Shield className="w-3 h-3 mr-1" />{user.role}</span>
              {user.is_banned && <span className="badge bg-red-100 text-red-600">Banned</span>}
            </div>
          </div>
        </div>
        <div className="space-y-2 text-sm text-gray-500">
          <div className="flex items-center gap-2"><Mail className="w-4 h-4" /><span>{user.email}</span></div>
          {user.faculty_id && <div className="flex items-center gap-2"><Building2 className="w-4 h-4" /><span>{faculties?.find(f => f.id === user.faculty_id)?.name ?? `Faculty #${user.faculty_id}`}</span></div>}
          <p className="text-xs text-gray-400">Member since {format(new Date(user.created_at), 'MMMM yyyy')}</p>
        </div>
      </div>
      <div className="card p-6">
        <h2 className="text-base font-semibold text-gray-900 mb-4">Update Profile</h2>
        <form onSubmit={handleSubmit((v) => mut.mutate(v))} className="space-y-4">
          <div><label className="block text-sm font-medium text-gray-700 mb-1">Email</label><input {...register('email')} type="email" className="input" />{errors.email && <p className="text-xs text-red-500 mt-1">{errors.email.message}</p>}</div>
          <div><label className="block text-sm font-medium text-gray-700 mb-1">Faculty</label>
            <select {...register('faculty_id', { valueAsNumber: true })} className="input"><option value="">— None —</option>{faculties?.map(f => <option key={f.id} value={f.id}>{f.name}</option>)}</select>
            {errors.faculty_id && <p className="text-xs text-red-500 mt-1">{errors.faculty_id.message}</p>}</div>
          {mut.error && <p className="text-sm text-red-500">{(mut.error as any).response?.data?.error || 'Update failed'}</p>}
          {saved && <div className="p-3 bg-green-50 border border-green-200 rounded-lg text-sm text-green-600">Profile updated successfully!</div>}
          <button type="submit" disabled={isSubmitting || mut.isPending} className="btn-primary">{mut.isPending ? 'Saving...' : 'Save Changes'}</button>
        </form>
      </div>
    </div>
  );
}
