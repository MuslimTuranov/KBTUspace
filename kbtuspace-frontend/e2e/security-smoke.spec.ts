import { test, expect } from '@playwright/test';
import {
  ADMIN_PASSWORD,
  API_URL,
  adminToken,
  auth,
  expectApiAvailable,
  getFaculties,
  loginApi,
  loginUi,
  registerUser,
} from './helpers';

function fakeJwt(payload: Record<string, unknown>) {
  const encode = (value: object) => Buffer.from(JSON.stringify(value)).toString('base64url');
  return `${encode({ alg: 'HS256', typ: 'JWT' })}.${encode(payload)}.bad-signature`;
}

test.describe.configure({ mode: 'serial' });

test.describe('Access and Security Journey', () => {
  test('student is blocked from admin and protected APIs enforce auth', async ({ page, request }) => {
    await expectApiAvailable(request);
    const [faculty] = await getFaculties(request);
    const student = await registerUser(request, faculty.id);
    const token = await loginApi(request, student.email, student.password);

    await loginUi(page, student.email, student.password);
    await page.goto('/admin');
    await expect(page).toHaveURL(/\/$/);
    await expect(page.getByRole('heading', { name: 'Feed' })).toBeVisible();

    const adminApi = await request.get(`${API_URL}/admin/users`, { headers: auth(token) });
    expect(adminApi.status()).toBe(403);
    const noToken = await request.get(`${API_URL}/profile`);
    expect(noToken.status()).toBe(401);
    const approve = await request.patch(`${API_URL}/admin/posts/1/approve`, { headers: auth(token) });
    expect(approve.status()).toBe(403);
  });

  test('expired and tampered tokens log the browser out', async ({ page }) => {
    await page.goto('/login');
    await page.evaluate(() => {
      const header = btoa(JSON.stringify({ alg: 'HS256', typ: 'JWT' }));
      const payload = btoa(JSON.stringify({ user_id: 1, role: 'student', exp: Math.floor(Date.now() / 1000) - 60 }));
      localStorage.setItem('token', `${header}.${payload}.expired`);
    });
    await page.goto('/');
    await expect(page).toHaveURL(/\/login$/);

    const futureExp = Math.floor(Date.now() / 1000) + 60 * 60;
    await page.evaluate((token) => localStorage.setItem('token', token), fakeJwt({ user_id: 1, role: 'student', exp: futureExp }));
    await page.goto('/');
    await expect(page).toHaveURL(/\/login$/);
  });
});

test.describe('Smoke and Regression @smoke', () => {
  test('frontend, core routes, session refresh, history, and backend public endpoints work', async ({ page, request }) => {
    await expectApiAvailable(request);
    const [faculty] = await getFaculties(request);
    const student = await registerUser(request, faculty.id);

    await page.goto('/login');
    await expect(page.getByRole('heading', { name: 'Welcome to UniHub' })).toBeVisible();
    await page.goto('/register');
    await expect(page.getByRole('heading', { name: 'Create an account' })).toBeVisible();

    await loginUi(page, student.email, student.password);
    await page.reload();
    await expect(page.getByRole('heading', { name: 'Feed' })).toBeVisible();
    await page.goto('/events');
    await expect(page.getByRole('heading', { name: 'Events' })).toBeVisible();
    await page.goto('/profile');
    await expect(page.getByRole('heading', { name: 'Profile' })).toBeVisible();

    await page.goBack();
    await expect(page.getByRole('heading', { name: 'Events' })).toBeVisible();
    await page.getByTitle('Logout').click();
    await expect(page).toHaveURL(/\/login$/);
    await page.goBack();
    await expect(page).toHaveURL(/\/login$/);

    const apiRoot = API_URL.replace('/api/v1', '');
    const swagger = await request.get(`${apiRoot}/swagger/index.html`);
    expect(swagger.ok()).toBeTruthy();

    const token = await loginApi(request, student.email, student.password);
    const upload = await request.post(`${API_URL}/uploads/images`, {
      headers: auth(token),
      multipart: {
        image: {
          name: 'smoke.png',
          mimeType: 'image/png',
          buffer: Buffer.from(
            'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+/p9sAAAAASUVORK5CYII=',
            'base64',
          ),
        },
      },
    });
    expect(upload.ok()).toBeTruthy();
    const asset = await request.get(`${apiRoot}${(await upload.json()).url}`);
    expect(asset.ok()).toBeTruthy();
  });

  test('admin route renders for admin role when credentials are configured', async ({ page, request }) => {
    test.skip(!ADMIN_PASSWORD, 'Set E2E_ADMIN_PASSWORD to run admin smoke');
    await adminToken(request);
    await loginUi(page, process.env.E2E_ADMIN_EMAIL ?? 'admin@kbtu.kz', ADMIN_PASSWORD);
    await page.goto('/admin');
    await expect(page.getByRole('heading', { name: 'Admin Panel' })).toBeVisible();
  });
});
