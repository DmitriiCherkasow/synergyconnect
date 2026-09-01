// Тип пользователя (совпадает с dto.UserResponse)
export interface User {
    id: string;
    email: string;
    first_name: string;
    last_name: string;
    avatar_url: string;
    bio: string;
    role: string; // "user" | "admin" | "super_admin"
    is_active: boolean;
    is_verified: boolean;
  }
  
  // Запрос на регистрацию (совпадает с dto.RegisterRequest)
  export interface RegisterRequest {
    email: string;
    first_name: string;
    last_name: string;
    password: string;
  }
  
  // Запрос на логин (совпадает с dto.LoginRequest)
  export interface LoginRequest {
    email: string;
    password: string;
  }
  
  // Ответ после успешного логина
  export interface AuthResponse {
    access_token: string;
    refresh_token: string;
    expires_in: number; // секунды
    user: User;
  }
  
  // Ответ при обновлении токена
  export interface RefreshTokenResponse {
    access_token: string;
    refresh_token: string;
    expires_in: number;
  }
  
  // Состояние авторизации в Store
  export interface AuthState {
    user: User | null;
    accessToken: string | null;
    refreshToken: string | null;
    isAuthenticated: boolean;
    isLoading: boolean;
  }