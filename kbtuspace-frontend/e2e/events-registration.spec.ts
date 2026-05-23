import { test, expect } from '@playwright/test';
import {
  ADMIN_PASSWORD,
  API_URL,
  adminToken,
  auth,
  createEvent,
  expectApiAvailable,
  futureDate,
  getFaculties,
  loginApi,
  loginUi,
  pastDate,
  profile,
  registerUser,
  unique,
} from './helpers';

test.describe.configure({ mode: 'serial' });

test.describe('Events Registration Journey', () => {
  test.skip(!ADMIN_PASSWORD, 'Set E2E_ADMIN_PASSWORD to run event registration tests');

  test('student registers, refreshes, cancels, re-registers, and sees capacity/faculty errors', async ({ page, request }) => {
    await expectApiAvailable(request);
    const faculties = await getFaculties(request);
    test.skip(faculties.length < 2, 'At least two seeded faculties are required for cross-faculty checks');

    const admin = await adminToken(request);
    const student = await registerUser(request, faculties[0].id);
    const studentToken = await loginApi(request, student.email, student.password);
    const registrationEvent = await createEvent(request, admin, {
      title: unique('Registration event'),
      description: 'Student can register, cancel, and register again.',
      event_date: futureDate(5),
      location: 'Registration Hall',
      capacity: 3,
      scope: 'faculty',
      faculty_id: faculties[0].id,
    });

    await loginUi(page, student.email, student.password);
    await page.goto(`/events/${registrationEvent.id}`);
    await expect(page.getByRole('heading', { name: registrationEvent.title })).toBeVisible();
    await expect(page.getByText(/0 registered/)).toBeVisible();

    await page.getByRole('button', { name: 'Register' }).click();
    await expect(page.getByRole('button', { name: 'Cancel Registration' })).toBeVisible();
    await expect(page.getByText(/1 registered/)).toBeVisible();
    await page.reload();
    await expect(page.getByRole('button', { name: 'Cancel Registration' })).toBeVisible();

    await page.getByRole('button', { name: 'Cancel Registration' }).click();
    await expect(page.getByRole('button', { name: 'Register' })).toBeVisible();
    await expect(page.getByText(/0 registered/)).toBeVisible();
    await page.getByRole('button', { name: 'Register' }).click();
    await expect(page.getByRole('button', { name: 'Cancel Registration' })).toBeVisible();

    const fullEvent = await createEvent(request, admin, {
      title: unique('Full event'),
      description: 'Event with one seat already taken.',
      event_date: futureDate(6),
      location: 'Small Room',
      capacity: 1,
      scope: 'faculty',
      faculty_id: faculties[0].id,
    });
    const otherStudent = await registerUser(request, faculties[0].id);
    const otherToken = await loginApi(request, otherStudent.email, otherStudent.password);
    const fill = await request.post(`${API_URL}/events/${fullEvent.id}/register`, { headers: auth(otherToken) });
    expect(fill.ok()).toBeTruthy();
    const fullRegister = await request.post(`${API_URL}/events/${fullEvent.id}/register`, { headers: auth(studentToken) });
    expect(fullRegister.status()).toBe(409);

    const crossFacultyEvent = await createEvent(request, admin, {
      title: unique('Cross faculty event'),
      description: 'Cross faculty registration should be forbidden.',
      event_date: futureDate(7),
      location: 'Other Faculty',
      capacity: 5,
      scope: 'faculty',
      faculty_id: faculties[1].id,
    });
    const forbidden = await request.post(`${API_URL}/events/${crossFacultyEvent.id}/register`, { headers: auth(studentToken) });
    expect(forbidden.status()).toBe(403);

    const studentUser = await profile(request, studentToken);
    expect(studentUser.faculty_id).toBe(faculties[0].id);
    const pastEventCreate = await request.post(`${API_URL}/events`, {
      headers: auth(admin),
      data: {
        title: unique('Past event'),
        description: 'Backend blocks creating past events, so past registration controls cannot be reached through normal UI setup.',
        event_date: pastDate(1),
        location: 'Past Room',
        capacity: 5,
        scope: 'faculty',
        faculty_id: faculties[0].id,
      },
    });
    expect(pastEventCreate.status()).toBe(400);
  });
});
