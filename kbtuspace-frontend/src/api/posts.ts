import api from './client';
import type { Post, CreatePostRequest } from '../types';
export const getPosts = async (params?: { global?: boolean; faculty_id?: number }) => { const res = await api.get<Post[]>('/posts', { params }); return res.data; };
export const getPost = async (id: number) => { const res = await api.get<Post>(`/posts/${id}`); return res.data; };
export const createPost = async (data: CreatePostRequest) => { const res = await api.post<Post>('/posts', data); return res.data; };
export const updatePost = async (id: number, data: Partial<CreatePostRequest>) => { const res = await api.put<{ message: string }>(`/posts/${id}`, data); return res.data; };
export const deletePost = async (id: number) => { const res = await api.delete<{ message: string }>(`/posts/${id}`); return res.data; };
export const pinPost = async (id: number, is_pinned: boolean) => { const res = await api.patch<{ message: string }>(`/posts/${id}/pin`, { is_pinned }); return res.data; };
