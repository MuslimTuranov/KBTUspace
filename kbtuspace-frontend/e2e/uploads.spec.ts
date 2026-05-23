import { test, expect } from '@playwright/test';
import {
  API_URL,
  auth,
  createPost,
  expectApiAvailable,
  getFaculties,
  loginApi,
  loginUi,
  registerUser,
  unique,
} from './helpers';

const tinyPng = Buffer.from(
  'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+/p9sAAAAASUVORK5CYII=',
  'base64',
);

test.describe.configure({ mode: 'serial' });

test.describe('Uploads Journey', () => {
  test('post image appears in feed/detail; invalid and oversized uploads fail cleanly', async ({ page, request }) => {
    await expectApiAvailable(request);
    const [faculty] = await getFaculties(request);
    const student = await registerUser(request, faculty.id);
    const token = await loginApi(request, student.email, student.password);
    const imagePostTitle = unique('Image post');

    await loginUi(page, student.email, student.password);
    await page.getByRole('button', { name: 'New Post' }).click();
    await page.getByLabel('Title').fill(imagePostTitle);
    await page.getByLabel('Content').fill('This post includes an uploaded image.');
    await page.getByLabel('Image (optional)').setInputFiles({
      name: 'tiny.png',
      mimeType: 'image/png',
      buffer: tinyPng,
    });
    await page.getByRole('button', { name: 'Create Post' }).click();
    await expect(page.getByRole('link', { name: imagePostTitle })).toBeVisible();
    await expect(page.locator(`img`).first()).toBeVisible();
    await page.getByRole('link', { name: imagePostTitle }).click();
    await expect(page.getByRole('heading', { name: imagePostTitle })).toBeVisible();
    await expect(page.locator('img')).toBeVisible();

    const broken = await createPost(request, token, {
      title: unique('Broken image post'),
      content: 'This post points at a removed image and should still render.',
      scope: 'faculty',
      faculty_id: faculty.id,
      image_url: '/uploads/does-not-exist-e2e.png',
    });
    await page.goto(`/posts/${broken.id}`);
    await expect(page.getByRole('heading', { name: broken.title })).toBeVisible();
    await expect(page.getByText('This post points at a removed image and should still render.')).toBeVisible();

    const unsupported = await request.post(`${API_URL}/uploads/images`, {
      headers: auth(token),
      multipart: {
        image: {
          name: 'notes.txt',
          mimeType: 'text/plain',
          buffer: Buffer.from('not an image'),
        },
      },
    });
    expect(unsupported.status()).toBe(400);

    const oversized = await request.post(`${API_URL}/uploads/images`, {
      headers: auth(token),
      multipart: {
        image: {
          name: 'big.png',
          mimeType: 'image/png',
          buffer: Buffer.alloc(5 * 1024 * 1024 + 1),
        },
      },
    });
    expect(oversized.status()).toBe(400);
  });
});
