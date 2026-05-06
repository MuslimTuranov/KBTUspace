import api from './client';
import type { CreateReportRequest, Report } from '../types';
export const createReport = async (data: CreateReportRequest) => { const res = await api.post<{ message: string; report: Report }>('/reports', data); return res.data; };
