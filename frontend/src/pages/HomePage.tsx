import { useAuth } from '@/hooks/useAuth';
import { useNavigate } from 'react-router-dom';
import toast from 'react-hot-toast';

export const HomePage = () => {
  const { user, logout, isAuthenticated } = useAuth();
  const navigate = useNavigate();

  const handleLogout = async () => {
    try {
      await logout();
      toast.success('Вы вышли из системы');
      navigate('/login');
    } catch (error) {
      toast.error('Ошибка при выходе');
      console.error('Logout error:', error);
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      {/* Шапка */}
      <header className="border-b border-gray-200 bg-white dark:border-gray-700 dark:bg-gray-800">
        <div className="mx-auto flex max-w-7xl items-center justify-between px-4 py-4 sm:px-6 lg:px-8">
          <h1 className="text-xl font-bold text-gray-900 dark:text-white">
            SynergyConnect
          </h1>
          <div className="flex items-center gap-4">
            {isAuthenticated && user && (
              <span className="text-sm text-gray-700 dark:text-gray-300">
                {user.first_name} {user.last_name}
              </span>
            )}
            <button
              onClick={handleLogout}
              className="rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2"
            >
              Выйти
            </button>
          </div>
        </div>
      </header>

      {/* Основной контент */}
      <main className="mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8">
        <div className="text-center">
          <h2 className="text-3xl font-bold text-gray-900 dark:text-white">
            Добро пожаловать в SynergyConnect!
          </h2>
          <p className="mt-4 text-lg text-gray-600 dark:text-gray-400">
            {isAuthenticated && user
              ? `Привет, ${user.first_name}! Рады тебя видеть.`
              : 'Пожалуйста, войдите в систему.'}
          </p>

          {/* Информация о пользователе */}
          {isAuthenticated && user && (
            <div className="mx-auto mt-8 max-w-md rounded-lg bg-white p-6 shadow dark:bg-gray-800">
              <h3 className="mb-4 text-lg font-semibold text-gray-900 dark:text-white">
                Ваш профиль
              </h3>
              <div className="space-y-2 text-left text-sm text-gray-700 dark:text-gray-300">
                <p>
                  <span className="font-medium">Email:</span> {user.email}
                </p>
                <p>
                  <span className="font-medium">Имя:</span> {user.first_name} {user.last_name}
                </p>
                <p>
                  <span className="font-medium">Роль:</span>{' '}
                  <span className="capitalize">{user.role}</span>
                </p>
                <p>
                  <span className="font-medium">Статус:</span>{' '}
                  {user.is_verified ? '✅ Подтверждён' : '⏳ Не подтверждён'}
                </p>
              </div>
            </div>
          )}
        </div>
      </main>
    </div>
  );
};