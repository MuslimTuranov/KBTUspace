import { useEffect } from 'react';
import { X } from 'lucide-react';
interface Props { title: string; onClose: () => void; children: React.ReactNode; size?: 'sm' | 'md' | 'lg'; }
export default function Modal({ title, onClose, children, size = 'md' }: Props) {
  useEffect(() => { const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose(); }; document.addEventListener('keydown', onKey); return () => document.removeEventListener('keydown', onKey); }, [onClose]);
  const widths = { sm: 'max-w-sm', md: 'max-w-lg', lg: 'max-w-2xl' };
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50" onClick={(e) => { if (e.target === e.currentTarget) onClose(); }}>
      <div className={`card w-full ${widths[size]} flex flex-col max-h-[90vh]`}>
        <div className="flex items-center justify-between p-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold text-gray-900">{title}</h2>
          <button onClick={onClose} className="btn-ghost p-1 rounded-md"><X className="w-5 h-5" /></button>
        </div>
        <div className="p-4 overflow-y-auto">{children}</div>
      </div>
    </div>
  );
}
