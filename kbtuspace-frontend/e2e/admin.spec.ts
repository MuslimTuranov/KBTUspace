import { test, expect } from '@playwright/test';
import {
  ADMIN_PASSWORD,
  API_URL,
  adminToken,
  auth,
  createEvent,
  createPost,
  expectApiAvailable,
  futureDate,
  getFaculties,
  loginApi,
  loginUi,
  profile,
  registerUser,
  unique,
  updateUser,
} from './helpers';

test.describe.configure({ mode: 'serial' });

test.describe('Admin Journey', () => {
  test.skip(!ADMIN_PASSWORD, 'Set E2E_ADMIN_PASSWORD to run admin journey tests');

  test('admin moderates content and manages users', async ({ page, request }) => {
    await expectApiAvailable(request);
    const faculties = await getFaculties(request);
    const admin = await adminToken(request);

    await loginUi(page, process.env.E2E_ADMIN_EMAIL ?? 'admin@kbtu.kz', ADMIN_PASSWORD);
    await page.goto('/admin');
    await expect(page.getByRole('heading', { name: 'Admin Panel' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Content Moderation' })).toBeVisible();

    const student = await registerUser(request, faculties[0].id);
    const studentToken = await loginApi(request, student.email, student.password);
    const pendingPost = await createPost(request, studentToken, {
      title: unique('Pending global post'),
      content: 'A global post waiting for admin approval.',
      scope: 'global',
    });
    expect(pendingPost.status).toBe('pending');

    const organizer = await registerUser(request, faculties[0].id);
    const organizerTokenBefore = await loginApi(request, organizer.email, organizer.password);
    const organizerUser = await profile(request, organizerTokenBefore);
    await updateUser(request, admin, organizerUser.id, { role: 'organizer' });
    const organizerToken = await loginApi(request, organizer.email, organizer.password);
    const pendingEvent = await createEvent(request, organizerToken, {
      title: unique('Pending global event'),
      description: 'A global event waiting for admin rejection.',
      event_date: futureDate(9),
      location: 'Moderation Room',
      capacity: 15,
      scope: 'global',
    });

    await page.reload();
    await expect(page.getByText(pendingPost.title)).toBeVisible();
    await expect(page.getByText(pendingEvent.title)).toBeVisible();

    const approvePost = await request.patch(`${API_URL}/admin/posts/${pendingPost.id}/approve`, { headers: auth(admin) });
    expect(approvePost.ok()).toBeTruthy();
    const rejectEvent = await request.patch(`${API_URL}/admin/events/${pendingEvent.id}/reject`, {
      headers: auth(admin),
      data: { reason: 'E2E rejection reason' },
    });
    expect(rejectEvent.ok()).toBeTruthy();

    const deleteMe = await createPost(request, studentToken, {
      title: unique('Pending delete post'),
      content: 'A pending global post that the admin deletes.',
      scope: 'global',
    });
    const deletePending = await request.delete(`${API_URL}/admin/posts/${deleteMe.id}`, { headers: auth(admin) });
    expect(deletePending.ok()).toBeTruthy();

    await page.getByRole('button', { name: 'Users' }).click();
    await expect(page.getByPlaceholder('Filter by email...')).toBeVisible();
    await page.getByPlaceholder('Filter by email...').fill(student.email);
    await expect(page.getByText(student.email)).toBeVisible();

    const studentUser = await profile(request, studentToken);
    const promoted = await updateUser(request, admin, studentUser.id, { role: 'organizer' });
    expect(promoted.role).toBe('organizer');
    const demoted = await updateUser(request, admin, studentUser.id, { role: 'student' });
    expect(demoted.role).toBe('student');

    const banned = await updateUser(request, admin, studentUser.id, { is_banned: true });
    expect(banned.is_banned).toBeTruthy();
    const bannedApi = await request.get(`${API_URL}/profile`, { headers: auth(studentToken) });
    expect(bannedApi.status()).toBe(403);
    const bannedLogin = await request.post(`${API_URL}/auth/login`, { data: { email: student.email, password: student.password } });
    expect(bannedLogin.status()).toBe(403);
    const unbanned = await updateUser(request, admin, studentUser.id, { is_banned: false });
    expect(unbanned.is_banned).toBeFalsy();

    const adminGlobalPost = await createPost(request, admin, {
      title: unique('Admin global post'),
      content: 'Admin global post is visible immediately.',
      scope: 'global',
    });
    expect(adminGlobalPost.status).toBe('approved');

    const adminFacultyPost = await createPost(request, admin, {
      title: unique('Admin faculty post'),
      content: 'Admin faculty post is visible in selected faculty feed.',
      scope: 'faculty',
      faculty_id: faculties[0].id,
    });
    expect(adminFacultyPost.status).toBe('approved');

    const adminGlobalEvent = await createEvent(request, admin, {
      title: unique('Admin global event'),
      description: 'Admin global event is visible immediately.',
      event_date: futureDate(14),
      location: 'Global Stage',
      capacity: 100,
      scope: 'global',
    });
    expect(adminGlobalEvent.status).toBe('approved');

    const adminFacultyEvent = await createEvent(request, admin, {
      title: unique('Admin faculty event'),
      description: 'Admin faculty event is visible for selected faculty.',
      event_date: futureDate(15),
      location: 'Faculty Stage',
      capacity: 100,
      scope: 'faculty',
      faculty_id: faculties[0].id,
    });
    expect(adminFacultyEvent.status).toBe('approved');
  });
});
