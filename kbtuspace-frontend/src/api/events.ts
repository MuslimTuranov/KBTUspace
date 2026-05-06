import api from './client';
import type { Event, CreateEventRequest } from '../types';
export const getEvents = async (params?: { global?: boolean; faculty_id?: number }) => { const res = await api.get<Event[]>('/events', { params }); return res.data; };
export const getEvent = async (id: number) => { const res = await api.get<Event>(`/events/${id}`); return res.data; };
export const createEvent = async (data: CreateEventRequest) => { const res = await api.post<Event>('/events', data); return res.data; };
export const updateEvent = async (id: number, data: Partial<CreateEventRequest>) => { const res = await api.put<{ message: string }>(`/events/${id}`, data); return res.data; };
export const deleteEvent = async (id: number) => { const res = await api.delete<{ message: string }>(`/events/${id}`); return res.data; };
export const registerForEvent = async (id: number) => { const res = await api.post<{ message: string }>(`/events/${id}/register`); return res.data; };
export const cancelEventRegistration = async (id: number) => { const res = await api.delete<{ message: string }>(`/events/${id}/register`); return res.data; };
