import api from './client';
export interface Faculty { id: number; name: string; }
export const getFaculties = async () => { const res = await api.get<Faculty[]>('/faculties'); return res.data; };
