import { Navigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import type { Role } from '../types';

interface Props { children: React.ReactNode; requiredRole?: Role; }

export default function ProtectedRoute({ children, requiredRole }: Props) {
  const { user, isLoading } = useAuth();
  if (isLoading) return <div className="min-h-screen flex items-center justify-center"><div className="w-8 h-8 border-4 border-blue-600 border-t-transparent rounded-full animate-spin" /></div>;
  if (!user) return <Navigate to="/login" replace />;
  if (user.is_banned) return <div className="min-h-screen flex items-center justify-center"><div className="text-center"><h1 className="text-2xl font-bold text-red-600 mb-2">Account Banned</h1><p className="text-gray-600">Your account has been suspended.</p></div></div>;
  if (requiredRole && user.role !== requiredRole) return <Navigate to="/" replace />;
  return <>{children}</>;
}
