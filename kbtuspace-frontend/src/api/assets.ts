import api from './client';

export function assetUrl(path: string | null | undefined): string | undefined {
  if (!path) return undefined;
  if (/^https?:\/\//i.test(path)) return path;

  const apiBase = api.defaults.baseURL ?? '';
  const origin = apiBase.replace(/\/api\/v\d+\/?$/, '');
  return `${origin}${path}`;
}
