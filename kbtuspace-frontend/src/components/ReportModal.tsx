import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useMutation } from '@tanstack/react-query';
import { createReport } from '../api/reports';
import type { ReportTargetType } from '../types';
import Modal from './Modal';

const schema = z.object({ reason: z.string().min(3).max(1000) });
type FormValues = z.infer<typeof schema>;

export default function ReportModal({ targetType, targetId, onClose }: { targetType: ReportTargetType; targetId: number; onClose: () => void }) {
  const { register, handleSubmit, formState: { errors } } = useForm<FormValues>({ resolver: zodResolver(schema) });
  const mut = useMutation({ mutationFn: (v: FormValues) => createReport({ target_type: targetType, target_id: targetId, reason: v.reason }), onSuccess: onClose });
  return (
    <Modal title={`Report ${targetType}`} onClose={onClose} size="sm">
      <form onSubmit={handleSubmit((v) => mut.mutate(v))} className="space-y-4">
        <div><label className="block text-sm font-medium text-gray-700 mb-1">Reason</label>
          <textarea {...register('reason')} rows={4} className="input resize-none" placeholder="Describe why you're reporting this content..." />
          {errors.reason && <p className="text-xs text-red-500 mt-1">{errors.reason.message}</p>}</div>
        {mut.error && <p className="text-sm text-red-500">{(mut.error as any).response?.data?.error || 'Failed to submit'}</p>}
        <div className="flex gap-2 justify-end"><button type="button" onClick={onClose} className="btn-secondary">Cancel</button><button type="submit" disabled={mut.isPending} className="btn-danger">{mut.isPending ? 'Submitting...' : 'Submit Report'}</button></div>
      </form>
    </Modal>
  );
}
