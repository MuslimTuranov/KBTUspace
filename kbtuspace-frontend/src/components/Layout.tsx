import { Link, NavLink, useNavigate } from 'react-router-dom';
import { LogOut, User, Newspaper, Calendar, Shield, BookOpen } from 'lucide-react';
import { useAuth } from '../context/AuthContext';

export default function Layout({ children }: { children: React.ReactNode }) {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const navLinkClass = ({ isActive }: { isActive: boolean }) => `flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium transition-colors ${isActive ? 'bg-blue-50 text-blue-700' : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900'}`;
  return (
    <div className="min-h-screen flex flex-col">
      <header className="bg-white border-b border-gray-200 sticky top-0 z-40">
        <div className="max-w-6xl mx-auto px-4 h-14 flex items-center justify-between">
          <Link to="/" className="flex items-center gap-2 font-bold text-lg text-blue-700"><BookOpen className="w-5 h-5" />UniHub</Link>
          <nav className="flex items-center gap-1">
            <NavLink to="/" end className={navLinkClass}><Newspaper className="w-4 h-4" /><span className="hidden sm:inline">Feed</span></NavLink>
            <NavLink to="/events" className={navLinkClass}><Calendar className="w-4 h-4" /><span className="hidden sm:inline">Events</span></NavLink>
            {user?.role === 'admin' && <NavLink to="/admin" className={navLinkClass}><Shield className="w-4 h-4" /><span className="hidden sm:inline">Admin</span></NavLink>}
          </nav>
          <div className="flex items-center gap-2">
            <NavLink to="/profile" className={navLinkClass}><User className="w-4 h-4" /><span className="hidden sm:inline max-w-[120px] truncate">{user?.email}</span></NavLink>
            <button onClick={() => { logout(); navigate('/login'); }} className="btn-ghost p-2 rounded-lg" title="Logout"><LogOut className="w-4 h-4" /></button>
          </div>
        </div>
      </header>
      <main className="flex-1 max-w-6xl mx-auto w-full px-4 py-6">{children}</main>
    </div>
  );
}
