import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, waitFor, fireEvent, cleanup } from '@testing-library/react';
import { AuthProvider, useAuth } from './AuthContext';
import { getProfile } from '../api/auth';
import type { User } from '../types';

vi.mock('../api/auth', () => ({
    getProfile: vi.fn(),
}));

const mockedGetProfile = vi.mocked(getProfile);

const user: User = {
    id: 1,
    email: 'student@kbtu.kz',
    role: 'student',
    faculty_id: 1,
    is_banned: false,
    created_at: '2026-05-12T00:00:00Z',
    updated_at: '2026-05-12T00:00:00Z',
};

function createToken(expOffsetSeconds: number) {
    const payload = {
        user_id: 1,
        role: 'student',
        faculty_id: 1,
        exp: Math.floor(Date.now() / 1000) + expOffsetSeconds,
    };

    return `header.${btoa(JSON.stringify(payload))}.signature`;
}

function TestComponent() {
    const { user, token, isLoading, login, logout } = useAuth();

    return (
        <div>
            <div data-testid="loading">{String(isLoading)}</div>
            <div data-testid="token">{token ?? 'no-token'}</div>
            <div data-testid="user">{user?.email ?? 'no-user'}</div>

            <button onClick={() => login(createToken(3600))}>login</button>
            <button onClick={logout}>logout</button>
        </div>
    );
}

describe('AuthContext', () => {
    beforeEach(() => {
        cleanup();
        localStorage.clear();
        vi.clearAllMocks();
    });

    afterEach(() => {
        cleanup();
    });

    it('loads token from localStorage', async () => {
        const token = createToken(3600);
        localStorage.setItem('token', token);

        mockedGetProfile.mockResolvedValue(user);

        const { getByTestId } = render(
            <AuthProvider>
                <TestComponent />
            </AuthProvider>
        );

        await waitFor(() => {
            expect(getByTestId('loading').textContent).toBe('false');
        });

        expect(getByTestId('token').textContent).toBe(token);
    });

    it('removes expired token', async () => {
        const token = createToken(-3600);
        localStorage.setItem('token', token);

        const { getByTestId } = render(
            <AuthProvider>
                <TestComponent />
            </AuthProvider>
        );

        await waitFor(() => {
            expect(getByTestId('loading').textContent).toBe('false');
        });

        expect(localStorage.getItem('token')).toBeNull();
        expect(getByTestId('token').textContent).toBe('no-token');
    });

    it('removes invalid token', async () => {
        localStorage.setItem('token', 'invalid-token');

        const { getByTestId } = render(
            <AuthProvider>
                <TestComponent />
            </AuthProvider>
        );

        await waitFor(() => {
            expect(getByTestId('loading').textContent).toBe('false');
        });

        expect(localStorage.getItem('token')).toBeNull();
        expect(getByTestId('token').textContent).toBe('no-token');
    });

    it('calls getProfile when token is valid', async () => {
        const token = createToken(3600);
        localStorage.setItem('token', token);

        mockedGetProfile.mockResolvedValue(user);

        const { getByTestId } = render(
            <AuthProvider>
                <TestComponent />
            </AuthProvider>
        );

        await waitFor(() => {
            expect(mockedGetProfile).toHaveBeenCalledTimes(1);
        });

        expect(getByTestId('user').textContent).toBe('student@kbtu.kz');
    });

    it('clears token when getProfile fails', async () => {
        const token = createToken(3600);
        localStorage.setItem('token', token);

        mockedGetProfile.mockRejectedValue(new Error('profile error'));

        const { getByTestId } = render(
            <AuthProvider>
                <TestComponent />
            </AuthProvider>
        );

        await waitFor(() => {
            expect(getByTestId('loading').textContent).toBe('false');
        });

        expect(localStorage.getItem('token')).toBeNull();
        expect(getByTestId('token').textContent).toBe('no-token');
    });

    it('login saves token and user', async () => {
        mockedGetProfile.mockResolvedValue(user);

        const { getByTestId, getByText } = render(
            <AuthProvider>
                <TestComponent />
            </AuthProvider>
        );

        await waitFor(() => {
            expect(getByTestId('loading').textContent).toBe('false');
        });

        fireEvent.click(getByText('login'));

        await waitFor(() => {
            expect(getByTestId('user').textContent).toBe('student@kbtu.kz');
        });

        expect(localStorage.getItem('token')).not.toBeNull();
        expect(getByTestId('token').textContent).not.toBe('no-token');
    });

    it('logout clears token and user', async () => {
        const token = createToken(3600);
        localStorage.setItem('token', token);

        mockedGetProfile.mockResolvedValue(user);

        const { getByTestId, getByText } = render(
            <AuthProvider>
                <TestComponent />
            </AuthProvider>
        );

        await waitFor(() => {
            expect(getByTestId('user').textContent).toBe('student@kbtu.kz');
        });

        fireEvent.click(getByText('logout'));

        expect(localStorage.getItem('token')).toBeNull();
        expect(getByTestId('token').textContent).toBe('no-token');
        expect(getByTestId('user').textContent).toBe('no-user');
    });
});