import { useQuery } from '@tanstack/react-query';
import { getFaculties, type Faculty } from '../api/faculties';
export function useFaculties() { return useQuery<Faculty[]>({ queryKey: ['faculties'], queryFn: getFaculties, staleTime: Infinity }); }
export function useFacultyName(facultyId: number | null | undefined): string {
  const { data } = useFaculties();
  if (!facultyId) return '';
  return data?.find(f => f.id === facultyId)?.name ?? `Faculty #${facultyId}`;
}
