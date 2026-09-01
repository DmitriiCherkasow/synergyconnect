import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { Link, useNavigate } from 'react-router-dom';
import toast from 'react-hot-toast';
import { EyeIcon, EyeSlashIcon } from '@heroicons/react/24/outline';
import { useAuth } from '@/hooks/useAuth';
import { registerSchema, type RegisterFormData } from '@/utils/validators';

// Функция проверки сложности пароля
const getPasswordStrength = (password: string) => {
  let score = 0;
  if (password.length >= 8) score++;
  if (/[A-Z]/.test(password)) score++;
  if (/[a-z]/.test(password)) score++;
  if (/\d/.test(password)) score++;
  return score;
};

export const RegisterPage = () => {
  const [isLoading, setIsLoading] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const navigate = useNavigate();
  const { register: registerUser } = useAuth();

  const {
    register,
    handleSubmit,
    watch,
    formState: { errors },
  } = useForm<RegisterFormData>({
    resolver: zodResolver(registerSchema),
    mode: 'onChange',
  });

  const password = watch('password', '');
  const strength = getPasswordStrength(password);

  const onSubmit = async (data: RegisterFormData) => {
    setIsLoading(true);
    try {
      await registerUser(data);
      toast.success('Регистрация успешна! Добро пожаловать!');
      navigate('/', { replace: true });
    } catch (error) {
      toast.error('Ошибка регистрации. Попробуйте снова.');
      console.error('Register error:', error);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50 px-4 py-12 dark:bg-gray-900">
      <div className="w-full max-w-md">
        <div className="rounded-2xl bg-white p-8 shadow-lg dark:bg-gray-800">
          {/* Заголовок */}
          <div className="mb-8 text-center">
            <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
              Регистрация в SynergyConnect
            </h1>
            <p className="mt-2 text-sm text-gray-600 dark:text-gray-400">
              Создайте новую учетную запись
            </p>
          </div>

          {/* Форма */}
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
            {/* Имя */}
            <div>
              <label
                htmlFor="first_name"
                className="block text-sm font-medium text-gray-700 dark:text-gray-300"
              >
                Имя
              </label>
              <input
                id="first_name"
                type="text"
                placeholder="Иван"
                {...register('first_name')}
                className={`mt-1 w-full rounded-lg border px-4 py-2.5 focus:outline-none focus:ring-2 dark:bg-gray-700 dark:text-white ${
                  errors.first_name
                    ? 'border-red-500 focus:ring-red-500'
                    : 'border-gray-300 focus:ring-primary-500 dark:border-gray-600'
                }`}
              />
              {errors.first_name && (
                <p className="mt-1 text-sm text-red-600">{errors.first_name.message}</p>
              )}
              <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
                От 1 до 50 символов
              </p>
            </div>

            {/* Фамилия */}
            <div>
              <label
                htmlFor="last_name"
                className="block text-sm font-medium text-gray-700 dark:text-gray-300"
              >
                Фамилия
              </label>
              <input
                id="last_name"
                type="text"
                placeholder="Иванов"
                {...register('last_name')}
                className={`mt-1 w-full rounded-lg border px-4 py-2.5 focus:outline-none focus:ring-2 dark:bg-gray-700 dark:text-white ${
                  errors.last_name
                    ? 'border-red-500 focus:ring-red-500'
                    : 'border-gray-300 focus:ring-primary-500 dark:border-gray-600'
                }`}
              />
              {errors.last_name && (
                <p className="mt-1 text-sm text-red-600">{errors.last_name.message}</p>
              )}
              <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
                От 1 до 50 символов
              </p>
            </div>

            {/* Email */}
            <div>
              <label
                htmlFor="email"
                className="block text-sm font-medium text-gray-700 dark:text-gray-300"
              >
                Email
              </label>
              <input
                id="email"
                type="email"
                placeholder="your@email.com"
                {...register('email')}
                className={`mt-1 w-full rounded-lg border px-4 py-2.5 focus:outline-none focus:ring-2 dark:bg-gray-700 dark:text-white ${
                  errors.email
                    ? 'border-red-500 focus:ring-red-500'
                    : 'border-gray-300 focus:ring-primary-500 dark:border-gray-600'
                }`}
              />
              {errors.email && (
                <p className="mt-1 text-sm text-red-600">{errors.email.message}</p>
              )}
              <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
                Введите корректный email
              </p>
            </div>

            {/* Пароль с кнопкой показа */}
            <div>
              <label
                htmlFor="password"
                className="block text-sm font-medium text-gray-700 dark:text-gray-300"
              >
                Пароль
              </label>
              <div className="relative mt-1">
                <input
                  id="password"
                  type={showPassword ? 'text' : 'password'}
                  placeholder="Минимум 8 символов"
                  {...register('password')}
                  className={`w-full rounded-lg border px-4 py-2.5 pr-10 focus:outline-none focus:ring-2 dark:bg-gray-700 dark:text-white ${
                    errors.password
                      ? 'border-red-500 focus:ring-red-500'
                      : 'border-gray-300 focus:ring-primary-500 dark:border-gray-600'
                  }`}
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
                >
                  {showPassword ? (
                    <EyeSlashIcon className="h-5 w-5" />
                  ) : (
                    <EyeIcon className="h-5 w-5" />
                  )}
                </button>
              </div>

              {/* Подсказки по паролю */}
              {password.length > 0 && (
                <div className="mt-2 space-y-1">
                  <p className="text-xs text-gray-500 dark:text-gray-400">
                    Требования к паролю:
                  </p>
                  <ul className="space-y-0.5 text-xs">
                    <li className={password.length >= 8 ? 'text-green-600' : 'text-gray-400'}>
                      {password.length >= 8 ? '✅' : '⬜'} Минимум 8 символов
                    </li>
                    <li className={/[A-Z]/.test(password) ? 'text-green-600' : 'text-gray-400'}>
                      {/[A-Z]/.test(password) ? '✅' : '⬜'} Заглавная буква (A-Z)
                    </li>
                    <li className={/[a-z]/.test(password) ? 'text-green-600' : 'text-gray-400'}>
                      {/[a-z]/.test(password) ? '✅' : '⬜'} Строчная буква (a-z)
                    </li>
                    <li className={/\d/.test(password) ? 'text-green-600' : 'text-gray-400'}>
                      {/\d/.test(password) ? '✅' : '⬜'} Цифра (0-9)
                    </li>
                  </ul>
                  {/* Индикатор сложности */}
                  <div className="mt-1 flex gap-1">
                    {[1, 2, 3, 4].map((level) => (
                      <div
                        key={level}
                        className={`h-1 flex-1 rounded-full transition ${
                          strength >= level
                            ? strength <= 2
                              ? 'bg-yellow-500'
                              : 'bg-green-500'
                            : 'bg-gray-200 dark:bg-gray-600'
                        }`}
                      />
                    ))}
                  </div>
                </div>
              )}

              {errors.password && (
                <p className="mt-1 text-sm text-red-600">{errors.password.message}</p>
              )}
            </div>

            {/* Кнопка регистрации */}
            <button
              type="submit"
              disabled={isLoading}
              className="w-full rounded-lg bg-primary-600 py-2.5 font-medium text-white transition hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
            >
              {isLoading ? 'Регистрация...' : 'Зарегистрироваться'}
            </button>
          </form>

          {/* Ссылка на вход */}
          <p className="mt-6 text-center text-sm text-gray-600 dark:text-gray-400">
            Уже есть аккаунт?{' '}
            <Link
              to="/login"
              className="font-medium text-primary-600 hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300"
            >
              Войти
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
};