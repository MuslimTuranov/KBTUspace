import api from './client';
import type { User } from '../types';
export interface LoginRequest { email: string; password: string; }
export interface RegisterRequest { email: string; password: string; }
export const login = async (data: LoginRequest) => { const res = await api.post<{ message: string; token: string }>('/auth/login', data); return res.data; };
export const register = async (data: RegisterRequest) => { const res = await api.post<{ message: string }>('/auth/register', data); return res.data; };
export const getProfile = async () => { const res = await api.get<User>('/profile'); return res.data; };
export const updateProfile = async (data: Partial<Pick<User, 'email' | 'faculty_id'>>) => { const res = await api.put<User>('/profile', data); return res.data; };
