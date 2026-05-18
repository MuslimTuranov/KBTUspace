import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { format } from 'date-fns';
import { User, Mail, Shield, Building2, Lock } from 'lucide-react';
import { changePassword } from '../api/auth';
import { getApiErrorMessage } from '../api/errors';
import { useAuth } from '../context/AuthContext';
import { useFaculties } from '../hooks/useFaculties';
import { useState } from 'react';

const schema = z.object({
  current_password: z.string().min(1, 'Current password is required'),
  new_password: z.string().min(6, 'New password must be at least 6 characters'),
  confirm_password: z.string().min(6, 'Confirm password is required'),
}).refine((d) => d.new_password === d.confirm_password, { message: "Passwords don't match", path: ['confirm_password'] });
type FormValues = z.infer<typeof schema>;
const roleBadgeClass: Record<string, string> = { student: 'bg-green-100 text-green-700', organizer: 'bg-blue-100 text-blue-700', admin: 'bg-purple-100 text-purple-700' };

export default function ProfilePage() {
  const { user, login } = useAuth();
  const { data: faculties } = useFaculties();
  const qc = useQueryClient();
  const [saved, setSaved] = useState(false);
  const { register, handleSubmit, reset, formState: { errors, isSubmitting } } = useForm<FormValues>({ resolver: zodResolver(schema) });
  const mut = useMutation({
    mutationFn: (v: FormValues) => changePassword({ current_password: v.current_password, new_password: v.new_password }),
    onSuccess: async () => {
      const token = localStorage.getItem('token');
      if (token) await login(token);
      qc.invalidateQueries({ queryKey: ['faculties'] });
      reset();
      setSaved(true); setTimeout(() => setSaved(false), 3000);
    },
  });
  if (!user) return null;

  const facultyName = faculties?.find(f => f.id === user.faculty_id)?.name ?? 'Not assigned';

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
          {user.role !== 'admin' && user.faculty_id && <div className="flex items-center gap-2"><Building2 className="w-4 h-4" /><span>{facultyName}</span></div>}
          <p className="text-xs text-gray-400">Member since {format(new Date(user.created_at), 'MMMM yyyy')}</p>
        </div>
      </div>
      <div className="card p-6">
        <h2 className="text-base font-semibold text-gray-900 mb-4">Change Password</h2>
        <form onSubmit={handleSubmit((v) => mut.mutate(v))} className="space-y-4">
          <div><label className="block text-sm font-medium text-gray-700 mb-1">Current password</label><input {...register('current_password')} type="password" role="textbox" className="input" />{errors.current_password && <p className="text-xs text-red-500 mt-1">{errors.current_password.message}</p>}</div>
          <div><label className="block text-sm font-medium text-gray-700 mb-1">New password</label><input {...register('new_password')} type="password" role="textbox" className="input" />{errors.new_password && <p className="text-xs text-red-500 mt-1">{errors.new_password.message}</p>}</div>
          <div><label className="block text-sm font-medium text-gray-700 mb-1">Confirm new password</label><input {...register('confirm_password')} type="password" role="textbox" className="input" />{errors.confirm_password && <p className="text-xs text-red-500 mt-1">{errors.confirm_password.message}</p>}</div>
          {mut.error && <p className="text-sm text-red-500">{getApiErrorMessage(mut.error, 'Update failed')}</p>}
          {saved && <div className="p-3 bg-green-50 border border-green-200 rounded-lg text-sm text-green-600">Password changed successfully!</div>}
          <button type="submit" disabled={isSubmitting || mut.isPending} className="btn-primary"><Lock className="w-4 h-4" />{mut.isPending ? 'Saving...' : 'Change Password'}</button>
        </form>
      </div>
    </div>
  );
}
