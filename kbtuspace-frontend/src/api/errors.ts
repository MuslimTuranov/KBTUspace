import { AxiosError } from 'axios';

interface ErrorResponse {
  error?: string;
}

export function getApiErrorMessage(error: unknown, fallback: string): string {
  if (error instanceof AxiosError) {
    const data = error.response?.data as ErrorResponse | undefined;
    return data?.error || fallback;
  }

  return fallback;
}
