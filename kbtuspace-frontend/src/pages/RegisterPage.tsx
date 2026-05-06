import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Link, useNavigate } from 'react-router-dom';
import { BookOpen } from 'lucide-react';
import { register as registerApi } from '../api/auth';
import { useState } from 'react';

const schema = z.object({ email: z.string().email('Invalid email'), password: z.string().min(6), confirmPassword: z.string() }).refine((d) => d.password === d.confirmPassword, { message: "Passwords don't match", path: ['confirmPassword'] });
type FormValues = z.infer<typeof schema>;

export default function RegisterPage() {
  const navigate = useNavigate();
  const [apiError, setApiError] = useState('');
  const { register, handleSubmit, formState: { errors, isSubmitting } } = useForm<FormValues>({ resolver: zodResolver(schema) });
  const onSubmit = async (values: FormValues) => {
    setApiError('');
    try { await registerApi({ email: values.email, password: values.password }); navigate('/login'); }
    catch (err: any) { setApiError(err.response?.data?.error || 'Registration failed'); }
  };
  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100 px-4">
      <div className="card w-full max-w-md p-8">
        <div className="text-center mb-8"><div className="inline-flex items-center justify-center w-14 h-14 bg-blue-600 rounded-2xl mb-4"><BookOpen className="w-7 h-7 text-white" /></div><h1 className="text-2xl font-bold text-gray-900">Create an account</h1><p className="text-gray-500 mt-1">Join the UniHub community</p></div>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div><label className="block text-sm font-medium text-gray-700 mb-1">Email</label><input {...register('email')} type="email" className="input" placeholder="you@kbtu.kz" />{errors.email && <p className="text-xs text-red-500 mt-1">{errors.email.message}</p>}</div>
          <div><label className="block text-sm font-medium text-gray-700 mb-1">Password</label><input {...register('password')} type="password" className="input" placeholder="Min. 6 characters" />{errors.password && <p className="text-xs text-red-500 mt-1">{errors.password.message}</p>}</div>
          <div><label className="block text-sm font-medium text-gray-700 mb-1">Confirm Password</label><input {...register('confirmPassword')} type="password" className="input" placeholder="••••••••" />{errors.confirmPassword && <p className="text-xs text-red-500 mt-1">{errors.confirmPassword.message}</p>}</div>
          {apiError && <div className="p-3 bg-red-50 border border-red-200 rounded-lg text-sm text-red-600">{apiError}</div>}
          <button type="submit" disabled={isSubmitting} className="btn-primary w-full justify-center py-2.5">{isSubmitting ? 'Creating account...' : 'Create Account'}</button>
        </form>
        <p className="text-center text-sm text-gray-500 mt-6">Already have an account? <Link to="/login" className="text-blue-600 hover:text-blue-700 font-medium">Sign in</Link></p>
      </div>
    </div>
  );
}
