import { expect, type APIRequestContext, type Page } from '@playwright/test';

export const API_URL = process.env.E2E_API_URL ?? process.env.VITE_API_URL ?? 'http://127.0.0.1:8080/api/v1';
export const ADMIN_EMAIL = process.env.E2E_ADMIN_EMAIL ?? 'admin@kbtu.kz';
export const ADMIN_PASSWORD = process.env.E2E_ADMIN_PASSWORD ?? '';

export type User = {
  id: number;
  email: string;
  role: 'student' | 'organizer' | 'admin';
  faculty_id: number | null;
  is_banned: boolean;
};

export type Faculty = { id: number; name: string };
export type Post = {
  id: number;
  title: string;
  content: string;
  status: string;
  scope: string;
  faculty_id: number | null;
  image_url?: string | null;
};
export type Event = Post & { event_date: string; capacity: number; current_count: number; is_registered?: boolean };

export function unique(prefix: string) {
  return `${prefix}-${Date.now()}-${Math.random().toString(16).slice(2, 8)}`;
}

export function kbtuEmail(prefix: string) {
  return `${unique(prefix)}@kbtu.kz`;
}

export async function expectApiAvailable(request: APIRequestContext) {
  const res = await request.get(API_URL.replace('/api/v1', '/ping'));
  expect(res.ok(), `API is not reachable at ${API_URL}`).toBeTruthy();
}

export async function getFaculties(request: APIRequestContext): Promise<Faculty[]> {
  const res = await request.get(`${API_URL}/faculties`);
  expect(res.ok()).toBeTruthy();
  const faculties = await res.json();
  expect(faculties.length, 'seeded faculties are required for e2e tests').toBeGreaterThan(0);
  return faculties;
}

export async function registerUser(request: APIRequestContext, facultyId: number, email = kbtuEmail('student'), password = 'Password123!') {
  const res = await request.post(`${API_URL}/auth/register`, { data: { email, password, faculty_id: facultyId } });
  expect([201, 409]).toContain(res.status());
  return { email, password, facultyId };
}

export async function loginApi(request: APIRequestContext, email: string, password: string) {
  const res = await request.post(`${API_URL}/auth/login`, { data: { email, password } });
  expect(res.ok(), `login failed for ${email}`).toBeTruthy();
  const body = await res.json();
  return body.token as string;
}

export async function profile(request: APIRequestContext, token: string): Promise<User> {
  const res = await request.get(`${API_URL}/profile`, { headers: auth(token) });
  expect(res.ok()).toBeTruthy();
  return res.json();
}

export function auth(token: string) {
  return { Authorization: `Bearer ${token}` };
}

export async function adminToken(request: APIRequestContext) {
  testAdminConfigured();
  return loginApi(request, ADMIN_EMAIL, ADMIN_PASSWORD);
}

export function testAdminConfigured() {
  expect(ADMIN_PASSWORD, 'Set E2E_ADMIN_PASSWORD to run admin/organizer e2e tests').not.toEqual('');
}

export async function updateUser(request: APIRequestContext, token: string, userId: number, data: Partial<Pick<User, 'role' | 'faculty_id' | 'is_banned'>>) {
  const res = await request.patch(`${API_URL}/admin/users/${userId}`, { headers: auth(token), data });
  expect(res.ok()).toBeTruthy();
  return res.json() as Promise<User>;
}

export async function createPost(request: APIRequestContext, token: string, data: Partial<Post> & { title: string; content: string }) {
  const res = await request.post(`${API_URL}/posts`, { headers: auth(token), data });
  expect(res.ok()).toBeTruthy();
  return (await res.json()).post as Post;
}

export async function updatePost(request: APIRequestContext, token: string, postId: number, data: Partial<Post> & { title: string; content: string }) {
  const res = await request.put(`${API_URL}/posts/${postId}`, { headers: auth(token), data });
  expect(res.ok()).toBeTruthy();
}

export async function createEvent(request: APIRequestContext, token: string, data: {
  title: string;
  description: string;
  event_date: string;
  location: string;
  capacity: number;
  scope?: 'faculty' | 'global';
  faculty_id?: number;
}) {
  const res = await request.post(`${API_URL}/events`, { headers: auth(token), data });
  expect(res.ok()).toBeTruthy();
  return (await res.json()).event as Event;
}

export async function approvePost(request: APIRequestContext, token: string, postId: number) {
  const res = await request.patch(`${API_URL}/admin/posts/${postId}/approve`, { headers: auth(token) });
  expect(res.ok()).toBeTruthy();
}

export async function approveEvent(request: APIRequestContext, token: string, eventId: number) {
  const res = await request.patch(`${API_URL}/admin/events/${eventId}/approve`, { headers: auth(token) });
  expect(res.ok()).toBeTruthy();
}

export async function loginUi(page: Page, email: string, password: string) {
  await page.goto('/login');
  await page.getByLabel('Email').fill(email);
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Sign in' }).click();
  await expect(page).toHaveURL(/\/$/);
  await expect(page.getByRole('heading', { name: 'Feed' })).toBeVisible();
}

export async function logoutUi(page: Page) {
  await page.getByTitle('Logout').click();
  await expect(page).toHaveURL(/\/login$/);
}

export async function registerUi(page: Page, email: string, password: string, facultyName?: string) {
  await page.goto('/register');
  await page.getByLabel('Email').fill(email);
  if (facultyName) {
    await page.getByLabel('Faculty').selectOption({ label: facultyName });
  } else {
    await page.getByLabel('Faculty').selectOption({ index: 1 });
  }
  await page.getByLabel('Password').fill(password);
  await page.getByLabel('Confirm Password').fill(password);
  await page.getByRole('button', { name: 'Create Account' }).click();
  await expect(page).toHaveURL(/\/login$/);
}

export function futureDate(days = 7) {
  return new Date(Date.now() + days * 24 * 60 * 60 * 1000).toISOString();
}

export function pastDate(days = 7) {
  return new Date(Date.now() - days * 24 * 60 * 60 * 1000).toISOString();
}
