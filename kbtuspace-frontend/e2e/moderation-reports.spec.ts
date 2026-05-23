import { test, expect } from '@playwright/test';
import {
  ADMIN_PASSWORD,
  API_URL,
  adminToken,
  auth,
  createPost,
  expectApiAvailable,
  getFaculties,
  loginApi,
  registerUser,
  unique,
} from './helpers';

test.describe.configure({ mode: 'serial' });

test.describe('Moderation and Reports Journey', () => {
  test.skip(!ADMIN_PASSWORD, 'Set E2E_ADMIN_PASSWORD to run report moderation tests');

  test('users report content and admin rejects/closes reports', async ({ request }) => {
    await expectApiAvailable(request);
    const [faculty] = await getFaculties(request);
    const admin = await adminToken(request);
    const userA = await registerUser(request, faculty.id);
    const userB = await registerUser(request, faculty.id);
    const tokenA = await loginApi(request, userA.email, userA.password);
    const tokenB = await loginApi(request, userB.email, userB.password);

    const firstPost = await createPost(request, tokenA, {
      title: unique('Reported post remains'),
      content: 'This post is reported once and remains after report rejection.',
      scope: 'faculty',
      faculty_id: faculty.id,
    });

    const report = await request.post(`${API_URL}/reports`, {
      headers: auth(tokenB),
      data: { target_type: 'post', target_id: firstPost.id, reason: 'Looks suspicious' },
    });
    expect(report.ok()).toBeTruthy();

    const duplicate = await request.post(`${API_URL}/reports`, {
      headers: auth(tokenB),
      data: { target_type: 'post', target_id: firstPost.id, reason: 'Still suspicious' },
    });
    expect(duplicate.status()).toBe(409);

    const ownReport = await request.post(`${API_URL}/reports`, {
      headers: auth(tokenA),
      data: { target_type: 'post', target_id: firstPost.id, reason: 'Own content report' },
    });
    expect(ownReport.status()).toBe(400);

    const pendingReports = await request.get(`${API_URL}/admin/reports`, { headers: auth(admin), params: { status: 'pending' } });
    expect(pendingReports.ok()).toBeTruthy();
    const pendingReport = (await pendingReports.json()).find((item: { target_title: string }) => item.target_title === firstPost.title);
    expect(pendingReport).toBeTruthy();

    const rejectReport = await request.patch(`${API_URL}/admin/reports/${pendingReport.id}/close`, {
      headers: auth(admin),
      data: { status: 'rejected', review_note: 'Content is acceptable' },
    });
    expect(rejectReport.ok()).toBeTruthy();
    const firstPostStillOpens = await request.get(`${API_URL}/posts/${firstPost.id}`, { headers: auth(tokenB) });
    expect(firstPostStillOpens.ok()).toBeTruthy();

    const secondPost = await createPost(request, tokenA, {
      title: unique('Reported post deleted'),
      content: 'This post is removed when the report is closed.',
      scope: 'faculty',
      faculty_id: faculty.id,
    });
    const secondReport = await request.post(`${API_URL}/reports`, {
      headers: auth(tokenB),
      data: { target_type: 'post', target_id: secondPost.id, reason: 'Please remove this' },
    });
    expect(secondReport.ok()).toBeTruthy();
    const secondReportId = (await secondReport.json()).report.id;

    const closeReport = await request.patch(`${API_URL}/admin/reports/${secondReportId}/close`, {
      headers: auth(admin),
      data: { status: 'closed', review_note: 'Removing content' },
    });
    expect(closeReport.ok()).toBeTruthy();
    const deletedPost = await request.get(`${API_URL}/posts/${secondPost.id}`, { headers: auth(tokenB) });
    expect(deletedPost.status()).toBe(404);
    const feed = await request.get(`${API_URL}/posts`, { headers: auth(tokenB), params: { faculty_id: faculty.id } });
    expect((await feed.json()).some((post: { id: number }) => post.id === secondPost.id)).toBeFalsy();
  });
});
