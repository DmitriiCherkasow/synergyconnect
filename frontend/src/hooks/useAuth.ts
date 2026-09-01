import { useAuthStore } from '@/store/authStore';
import { useCallback } from 'react';

export const useAuth = () => {
  const {
    user,
    accessToken,
    refreshToken,
    isAuthenticated,
    isLoading,
    login,
    register,
    logout,
    refreshTokens,
    setUser,
    reset,
  } = useAuthStore();

  // Проверка роли
  const hasRole = useCallback(
    (role: string | string[]) => {
      if (!user) return false;
      if (Array.isArray(role)) {
        return role.includes(user.role);
      }
      return user.role === role;
    },
    [user]
  );

  // Проверка, является ли пользователь админом
  const isAdmin = useCallback(() => {
    return hasRole(['admin', 'super_admin']);
  }, [hasRole]);

  // Проверка, является ли пользователь суперадмином
  const isSuperAdmin = useCallback(() => {
    return hasRole('super_admin');
  }, [hasRole]);

  return {
    // Состояние
    user,
    accessToken,
    refreshToken,
    isAuthenticated,
    isLoading,
    // Действия
    login,
    register,
    logout,
    refreshTokens,
    setUser,
    reset,
    // Вспомогательные методы
    hasRole,
    isAdmin,
    isSuperAdmin,
  };
};