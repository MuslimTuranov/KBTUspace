import { describe, it, expect, beforeEach, vi } from 'vitest';
import { AxiosError, type InternalAxiosRequestConfig } from 'axios';
import api from './client';
import { getApiErrorMessage } from './errors';
import { assetUrl } from './assets';

describe('API client', () => {
    beforeEach(() => {
        localStorage.clear();
        vi.restoreAllMocks();
    });

    it('adds Authorization Bearer header', async () => {
        localStorage.setItem('token', 'test-token');

        const requestHandlers = api.interceptors.request.handlers;
        expect(requestHandlers).toBeDefined();

        const interceptor = requestHandlers![0]?.fulfilled;
        expect(interceptor).toBeDefined();

        const config = await interceptor?.({
            headers: {} as InternalAxiosRequestConfig['headers'],
        } as InternalAxiosRequestConfig);

        expect(config?.headers.Authorization).toBe('Bearer test-token');
    });

    it('removes token and redirects on 401', async () => {
        localStorage.setItem('token', 'token');

        const removeSpy = vi.spyOn(Storage.prototype, 'removeItem');

        Object.defineProperty(window, 'location', {
            writable: true,
            value: { href: '' },
        });

        const responseHandlers = api.interceptors.response.handlers;
        expect(responseHandlers).toBeDefined();

        const interceptor = responseHandlers![0]?.rejected;
        expect(interceptor).toBeDefined();

        const error = {
            response: {
                status: 401,
            },
        };

        await expect(interceptor?.(error)).rejects.toEqual(error);

        expect(removeSpy).toHaveBeenCalledWith('token');
        expect(window.location.href).toBe('/login');
    });

    it('getApiErrorMessage returns axios error message', () => {
        const error = new AxiosError('Request failed');

        error.response = {
            data: { error: 'invalid credentials' },
            status: 400,
            statusText: 'Bad Request',
            headers: {},
            config: {} as InternalAxiosRequestConfig,
        };

        expect(getApiErrorMessage(error, 'fallback'))
            .toBe('invalid credentials');
    });

    it('getApiErrorMessage returns fallback', () => {
        const error = new AxiosError('Request failed');

        expect(getApiErrorMessage(error, 'fallback'))
            .toBe('fallback');
    });

    it('getApiErrorMessage returns fallback for non axios error', () => {
        expect(
            getApiErrorMessage(new Error('oops'), 'fallback')
        ).toBe('fallback');
    });

    it('assetUrl keeps absolute URL', () => {
        const url = 'https://cdn.test.com/image.png';

        expect(assetUrl(url)).toBe(url);
    });

    it('assetUrl builds uploads path from API origin', () => {
        api.defaults.baseURL =
            'http://localhost:8080/api/v1';

        expect(assetUrl('/uploads/x.png'))
            .toBe('http://localhost:8080/uploads/x.png');
    });
});