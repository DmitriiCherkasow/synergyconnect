import { z } from 'zod';

// Схема для регистрации
export const registerSchema = z.object({
  email: z.string().email('Введите корректный email'),
  first_name: z.string().min(1, 'Имя обязательно').max(50, 'Имя слишком длинное'),
  last_name: z.string().min(1, 'Фамилия обязательна').max(50, 'Фамилия слишком длинная'),
  password: z
    .string()
    .min(8, 'Пароль должен содержать минимум 8 символов')
    .regex(/[A-Z]/, 'Пароль должен содержать хотя бы одну заглавную букву')
    .regex(/[a-z]/, 'Пароль должен содержать хотя бы одну строчную букву')
    .regex(/\d/, 'Пароль должен содержать хотя бы одну цифру'),
});

// Схема для логина
export const loginSchema = z.object({
  email: z.string().email('Введите корректный email'),
  password: z.string().min(1, 'Введите пароль'),
});

// Типы из схем
export type RegisterFormData = z.infer<typeof registerSchema>;
export type LoginFormData = z.infer<typeof loginSchema>;