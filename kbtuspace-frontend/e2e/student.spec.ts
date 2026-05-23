import { test, expect } from '@playwright/test';
import {
  API_URL,
  auth,
  expectApiAvailable,
  getFaculties,
  kbtuEmail,
  loginApi,
  loginUi,
  logoutUi,
  registerUser,
  registerUi,
  unique,
} from './helpers';

test.describe.configure({ mode: 'serial' });

test.describe('Student Journey', () => {
  test('student can register, use feeds, manage own post, change password, and log back in', async ({ page, request }) => {
    await expectApiAvailable(request);
    const faculties = await getFaculties(request);
    const email = kbtuEmail('student-e2e');
    const oldPassword = 'Password123!';
    const newPassword = 'NewPassword123!';
    const title = unique('Faculty post');
    const editedTitle = `${title} edited`;

    await page.goto('/');
    await expect(page).toHaveURL(/\/login$/);

    await page.goto('/register');
    await page.getByLabel('Email').fill(`${unique('outsider')}@gmail.com`);
    await page.getByLabel('Faculty').selectOption({ label: faculties[0].name });
    await page.getByLabel('Password').fill(oldPassword);
    await page.getByLabel('Confirm Password').fill(oldPassword);
    await page.getByRole('button', { name: 'Create Account' }).click();
    await expect(page.getByText(/kbtu|domain|email/i)).toBeVisible();

    await registerUi(page, email, oldPassword, faculties[0].name);
    await loginUi(page, email, oldPassword);

    await expect(page.getByRole('heading', { name: 'Feed' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Faculty' })).toBeVisible();
    await page.getByRole('button', { name: 'Global' }).click();
    await expect(page.getByText(/No posts yet|Global/)).toBeVisible();
    await page.getByRole('button', { name: 'Faculty' }).click();

    await page.getByRole('button', { name: 'New Post' }).click();
    await page.getByLabel('Title').fill(title);
    await page.getByLabel('Content').fill('This is a faculty post created by the student e2e journey.');
    await page.getByLabel('Scope').selectOption('faculty');
    await page.getByRole('button', { name: 'Create Post' }).click();
    await expect(page.getByRole('link', { name: title })).toBeVisible();
    await expect(page.getByTitle('Pin')).toHaveCount(0);

    await page.getByRole('link', { name: title }).click();
    await expect(page.getByRole('heading', { name: title })).toBeVisible();
    await page.locator('button:has(svg.lucide-pencil)').click();
    await page.getByLabel('Title').fill(editedTitle);
    await page.getByRole('button', { name: 'Save Changes' }).click();
    await expect(page.getByRole('heading', { name: editedTitle })).toBeVisible();

    page.once('dialog', (dialog) => dialog.accept());
    await page.locator('button:has(svg.lucide-trash-2)').click();
    await expect(page).toHaveURL(/\/$/);
    await expect(page.getByRole('link', { name: editedTitle })).toHaveCount(0);

    await page.getByRole('link', { name: email }).click();
    await page.getByLabel('Current password').fill(oldPassword);
    await page.getByLabel('New password').fill(newPassword);
    await page.getByLabel('Confirm new password').fill(newPassword);
    await page.getByRole('button', { name: 'Change Password' }).click();
    await expect(page.getByText('Password changed successfully!')).toBeVisible();

    await logoutUi(page);
    await page.getByLabel('Email').fill(email);
    await page.getByLabel('Password').fill(oldPassword);
    await page.getByRole('button', { name: 'Sign in' }).click();
    await expect(page.getByText(/Invalid email or password|Invalid credentials/i)).toBeVisible();

    await loginUi(page, email, newPassword);
  });

  test('student cannot pin or mutate content owned by someone else via API', async ({ request }) => {
    const [faculty] = await getFaculties(request);
    const owner = await registerUser(request, faculty.id);
    const other = await registerUser(request, faculty.id);
    const ownerToken = await loginApi(request, owner.email, owner.password);
    const otherToken = await loginApi(request, other.email, other.password);

    const created = await request.post(`${API_URL}/posts`, {
      headers: auth(ownerToken),
      data: {
        title: unique('Owner post'),
        content: 'Post owned by another student for authorization checks.',
        scope: 'faculty',
        faculty_id: faculty.id,
      },
    });
    expect(created.ok()).toBeTruthy();
    const postId = (await created.json()).post.id;

    const pin = await request.patch(`${API_URL}/posts/${postId}/pin`, { headers: auth(otherToken), data: { is_pinned: true } });
    expect(pin.status()).toBe(403);

    const edit = await request.put(`${API_URL}/posts/${postId}`, {
      headers: auth(otherToken),
      data: { title: 'Forbidden edit', content: 'Non author should not be able to edit.', scope: 'faculty', faculty_id: faculty.id },
    });
    expect(edit.status()).toBe(403);

    const del = await request.delete(`${API_URL}/posts/${postId}`, { headers: auth(otherToken) });
    expect(del.status()).toBe(403);
  });
});
