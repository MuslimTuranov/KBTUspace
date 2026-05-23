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
  profile,
  registerUser,
  unique,
  updateUser,
} from './helpers';

test.describe.configure({ mode: 'serial' });

test.describe('Organizer Journey', () => {
  test.skip(!ADMIN_PASSWORD, 'Set E2E_ADMIN_PASSWORD to run admin-backed organizer tests');

  test('admin promotes student; organizer creates and manages faculty/global events', async ({ page, request }) => {
    await expectApiAvailable(request);
    const faculties = await getFaculties(request);
    test.skip(faculties.length < 2, 'At least two seeded faculties are required for cross-faculty checks');

    const admin = await adminToken(request);
    const organizer = await registerUser(request, faculties[0].id);
    const organizerTokenBefore = await loginApi(request, organizer.email, organizer.password);
    const organizerUser = await profile(request, organizerTokenBefore);
    await updateUser(request, admin, organizerUser.id, { role: 'organizer' });
    const organizerToken = await loginApi(request, organizer.email, organizer.password);

    await loginUi(page, organizer.email, organizer.password);
    await page.getByRole('link', { name: 'Events' }).click();
    await expect(page.getByRole('button', { name: 'New Event' })).toBeVisible();

    const facultyTitle = unique('Faculty event');
    await page.getByRole('button', { name: 'New Event' }).click();
    await page.getByLabel('Title').fill(facultyTitle);
    await page.getByLabel('Description').fill('Organizer creates a faculty event from the UI.');
    await page.getByLabel('Date & Time').fill(new Date(Date.now() + 8 * 24 * 60 * 60 * 1000).toISOString().slice(0, 16));
    await page.getByLabel('Capacity').fill('5');
    await page.getByLabel('Location').fill('KBTU Hall');
    await page.getByLabel('Scope').selectOption('faculty');
    await page.getByRole('button', { name: 'Create Event' }).click();
    await expect(page.getByRole('link', { name: facultyTitle })).toBeVisible();

    const student = await registerUser(request, faculties[0].id);
    const studentToken = await loginApi(request, student.email, student.password);
    const events = await request.get(`${API_URL}/events`, { headers: auth(studentToken), params: { faculty_id: faculties[0].id } });
    expect(events.ok()).toBeTruthy();
    const facultyEvent = (await events.json()).find((event: { title: string }) => event.title === facultyTitle);
    expect(facultyEvent).toBeTruthy();

    const register = await request.post(`${API_URL}/events/${facultyEvent.id}/register`, { headers: auth(studentToken) });
    expect(register.ok()).toBeTruthy();
    const studentUser = await profile(request, studentToken);
    const attendance = await request.patch(`${API_URL}/events/${facultyEvent.id}/attendance/${studentUser.id}`, { headers: auth(organizerToken) });
    expect(attendance.ok()).toBeTruthy();

    const globalTitle = unique('Global pending event');
    const globalEvent = await createEvent(request, organizerToken, {
      title: globalTitle,
      description: 'Organizer global event should wait for admin approval.',
      event_date: futureDate(10),
      location: 'Main Auditorium',
      capacity: 20,
      scope: 'global',
    });
    expect(globalEvent.status).toBe('pending');

    const globalBefore = await request.get(`${API_URL}/events`, { headers: auth(studentToken), params: { global: true } });
    expect(globalBefore.ok()).toBeTruthy();
    expect((await globalBefore.json()).some((event: { id: number }) => event.id === globalEvent.id)).toBeFalsy();

    const pending = await request.get(`${API_URL}/admin/moderation/global-content`, { headers: auth(admin), params: { type: 'events' } });
    expect(pending.ok()).toBeTruthy();
    expect((await pending.json()).some((event: { id: number }) => event.id === globalEvent.id)).toBeTruthy();

    const approve = await request.patch(`${API_URL}/admin/events/${globalEvent.id}/approve`, { headers: auth(admin) });
    expect(approve.ok()).toBeTruthy();
    const globalAfter = await request.get(`${API_URL}/events`, { headers: auth(studentToken), params: { global: true } });
    expect((await globalAfter.json()).some((event: { id: number }) => event.id === globalEvent.id)).toBeTruthy();

    const otherFacultyEvent = await createEvent(request, admin, {
      title: unique('Other faculty event'),
      description: 'Admin-created other faculty event for forbidden organizer checks.',
      event_date: futureDate(12),
      location: 'Other Campus',
      capacity: 10,
      scope: 'faculty',
      faculty_id: faculties[1].id,
    });
    const forbidden = await request.put(`${API_URL}/events/${otherFacultyEvent.id}`, {
      headers: auth(organizerToken),
      data: {
        title: 'Forbidden organizer edit',
        description: 'Organizer cannot manage another faculty event.',
        event_date: futureDate(13),
        location: 'Blocked',
        capacity: 10,
        scope: 'faculty',
        faculty_id: faculties[1].id,
      },
    });
    expect(forbidden.status()).toBe(403);
  });
});
