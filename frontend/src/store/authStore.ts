import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { User, AuthState } from '@/types/user';
import { login as apiLogin, register as apiRegister, logout as apiLogout, refreshToken as apiRefresh } from '@/api/auth';
import { setTokens, clearTokens } from '@/api/client';

interface AuthStore extends AuthState {
  // Действия
  login: (email: string, password: string) => Promise<void>;
  register: (data: { email: string; first_name: string; last_name: string; password: string }) => Promise<void>;
  logout: () => Promise<void>;
  refreshTokens: () => Promise<void>;  // Переименовали с refreshToken на refreshTokens
  setUser: (user: User) => void;
  reset: () => void;
}

export const useAuthStore = create<AuthStore>()(
  persist(
    (set, get) => ({
      // Начальное состояние
      user: null,
      accessToken: null,
      refreshToken: null,
      isAuthenticated: false,
      isLoading: false,

      // Вход
      login: async (email: string, password: string) => {
        set({ isLoading: true });
        try {
          const response = await apiLogin({ email, password });
          const { access_token, refresh_token, user } = response;
          
          set({
            user,
            accessToken: access_token,
            refreshToken: refresh_token,
            isAuthenticated: true,
            isLoading: false,
          });
          
          setTokens(access_token, refresh_token);
          localStorage.setItem('accessToken', access_token);
          localStorage.setItem('refreshToken', refresh_token);
        } catch (error) {
          set({ isLoading: false });
          throw error;
        }
      },

      // Регистрация
      register: async (data) => {
        set({ isLoading: true });
        try {
          const response = await apiRegister(data);
          const { access_token, refresh_token, user } = response;
          
          set({
            user,
            accessToken: access_token,
            refreshToken: refresh_token,
            isAuthenticated: true,
            isLoading: false,
          });
          
          setTokens(access_token, refresh_token);
          localStorage.setItem('accessToken', access_token);
          localStorage.setItem('refreshToken', refresh_token);
        } catch (error) {
          set({ isLoading: false });
          throw error;
        }
      },

      // Выход
      logout: async () => {
        set({ isLoading: true });
        try {
          await apiLogout();
        } catch (error) {
          console.warn('Logout error:', error);
        } finally {
          clearTokens();
          localStorage.removeItem('accessToken');
          localStorage.removeItem('refreshToken');
          set({
            user: null,
            accessToken: null,
            refreshToken: null,
            isAuthenticated: false,
            isLoading: false,
          });
        }
      },

      // Обновление токена (переименовано с refreshToken на refreshTokens)
      refreshTokens: async () => {
        const { refreshToken: currentRefreshToken } = get();
        if (!currentRefreshToken) return;

        try {
          const response = await apiRefresh(currentRefreshToken);
          const { access_token, refresh_token } = response;
          
          set({
            accessToken: access_token,
            refreshToken: refresh_token,
          });
          
          setTokens(access_token, refresh_token);
          localStorage.setItem('accessToken', access_token);
          localStorage.setItem('refreshToken', refresh_token);
        } catch (error) {
          console.error('Failed to refresh token:', error);
          await get().logout();
        }
      },

      // Установить пользователя
      setUser: (user: User) => {
        set({ user });
      },

      // Сброс состояния
      reset: () => {
        clearTokens();
        localStorage.removeItem('accessToken');
        localStorage.removeItem('refreshToken');
        set({
          user: null,
          accessToken: null,
          refreshToken: null,
          isAuthenticated: false,
          isLoading: false,
        });
      },
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        user: state.user,
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
);