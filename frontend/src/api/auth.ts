import { apiClient, setTokens, clearTokens } from './client';
import type { 
  RegisterRequest, 
  LoginRequest, 
  AuthResponse, 
  RefreshTokenResponse 
} from '@/types/user';

/**
 * Регистрация нового пользователя
 */
export const register = async (data: RegisterRequest): Promise<AuthResponse> => {
  const response = await apiClient.post<AuthResponse>('/auth/register', data);
  return response.data;
};

/**
 * Вход в систему
 */
export const login = async (data: LoginRequest): Promise<AuthResponse> => {
  const response = await apiClient.post<AuthResponse>('/auth/login', data);
  
  // Сохраняем токены
  const { access_token, refresh_token } = response.data;
  setTokens(access_token, refresh_token);
  
  return response.data;
};

/**
 * Обновление токена доступа
 */
export const refreshToken = async (refreshToken: string): Promise<RefreshTokenResponse> => {
  const response = await apiClient.post<RefreshTokenResponse>('/auth/refresh', {
    refresh_token: refreshToken,
  });
  
  // Обновляем токены
  const { access_token, refresh_token } = response.data;
  setTokens(access_token, refresh_token);
  
  return response.data;
};

/**
 * Выход из системы
 */
export const logout = async (): Promise<void> => {
  try {
    await apiClient.post('/auth/logout');
  } catch (error) {
    // Даже если сервер вернул ошибку, всё равно очищаем локальные токены
    console.warn('Logout error:', error);
  } finally {
    clearTokens();
  }
};

/**
 * Проверка текущего статуса авторизации (можно использовать для защиты маршрутов)
 */
export const checkAuth = (): boolean => {
  // Временно проверяем наличие токена в localStorage
  // Позже заменим на Zustand store
  const token = localStorage.getItem('accessToken');
  return !!token;
};