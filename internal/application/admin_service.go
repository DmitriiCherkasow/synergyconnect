package application

import (
	"context"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
	"github.com/google/uuid"
)

// AdminService — сервис для административных функций
type AdminService struct {
	userRepo    domain.UserRepository
	postRepo    domain.PostRepository
	commentRepo domain.CommentRepository
	projectRepo domain.ProjectRepository
	vacancyRepo domain.VacancyRepository
}

// NewAdminService создает новый сервис администратора
func NewAdminService(
	userRepo domain.UserRepository,
	postRepo domain.PostRepository,
	commentRepo domain.CommentRepository,
	projectRepo domain.ProjectRepository,
	vacancyRepo domain.VacancyRepository,
) *AdminService {
	return &AdminService{
		userRepo:    userRepo,
		postRepo:    postRepo,
		commentRepo: commentRepo,
		projectRepo: projectRepo,
		vacancyRepo: vacancyRepo,
	}
}

// ============================================
// Управление пользователями
// ============================================

// GetUsers возвращает список пользователей
func (s *AdminService) GetUsers(ctx context.Context, filter domain.UserFilter) ([]domain.User, int64, error) {
	return s.userRepo.List(ctx, filter)
}

// GetUserByID возвращает пользователя по ID
func (s *AdminService) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.userRepo.FindByID(ctx, id)
}

// BlockUser блокирует пользователя
func (s *AdminService) BlockUser(ctx context.Context, actorID, userID uuid.UUID) error {
	actor, err := s.userRepo.FindByID(ctx, actorID)
	if err != nil {
		return err
	}
	if actor == nil || !actor.CanManageUsers() {
		return domain.ErrInsufficientPermissions
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return domain.ErrUserNotFound
	}

	if user.IsSuperAdmin() {
		return domain.ErrCannotChangeSuperAdmin
	}

	user.IsActive = false
	return s.userRepo.Update(ctx, user)
}

// UnblockUser разблокирует пользователя
func (s *AdminService) UnblockUser(ctx context.Context, actorID, userID uuid.UUID) error {
	actor, err := s.userRepo.FindByID(ctx, actorID)
	if err != nil {
		return err
	}
	if actor == nil || !actor.CanManageUsers() {
		return domain.ErrInsufficientPermissions
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return domain.ErrUserNotFound
	}

	user.IsActive = true
	return s.userRepo.Update(ctx, user)
}

// DeleteUser удаляет пользователя
func (s *AdminService) DeleteUser(ctx context.Context, actorID, userID uuid.UUID) error {
	actor, err := s.userRepo.FindByID(ctx, actorID)
	if err != nil {
		return err
	}
	if actor == nil || !actor.CanManageUsers() {
		return domain.ErrInsufficientPermissions
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return domain.ErrUserNotFound
	}

	if user.IsSuperAdmin() {
		return domain.ErrCannotDeleteSuperAdmin
	}

	return s.userRepo.Delete(ctx, userID)
}

// PromoteToAdmin повышает пользователя до администратора
func (s *AdminService) PromoteToAdmin(ctx context.Context, actorID, userID uuid.UUID) error {
	actor, err := s.userRepo.FindByID(ctx, actorID)
	if err != nil {
		return err
	}
	if actor == nil || !actor.CanAssignAdmin() {
		return domain.ErrInsufficientPermissions
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return domain.ErrUserNotFound
	}

	if user.IsSuperAdmin() {
		return domain.ErrCannotChangeSuperAdmin
	}

	if user.Role == domain.RoleAdmin {
		return domain.ErrUserAlreadyAdmin
	}

	return s.userRepo.PromoteToAdmin(ctx, userID)
}

// DemoteFromAdmin понижает администратора до студента
func (s *AdminService) DemoteFromAdmin(ctx context.Context, actorID, userID uuid.UUID) error {
	actor, err := s.userRepo.FindByID(ctx, actorID)
	if err != nil {
		return err
	}
	if actor == nil || !actor.CanAssignAdmin() {
		return domain.ErrInsufficientPermissions
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return domain.ErrUserNotFound
	}

	if user.IsSuperAdmin() {
		return domain.ErrCannotChangeSuperAdmin
	}

	if user.Role != domain.RoleAdmin {
		return domain.ErrUserNotAdmin
	}

	return s.userRepo.DemoteFromAdmin(ctx, userID)
}

// GetAdmins возвращает всех администраторов
func (s *AdminService) GetAdmins(ctx context.Context) ([]domain.User, error) {
	return s.userRepo.FindAdmins(ctx)
}

// GetSuperAdmin возвращает суперадмина
func (s *AdminService) GetSuperAdmin(ctx context.Context) (*domain.User, error) {
	return s.userRepo.FindSuperAdmin(ctx)
}

// ============================================
// Управление контентом (посты и комментарии)
// ============================================

// GetPosts возвращает список постов
func (s *AdminService) GetPosts(ctx context.Context, filter domain.PostFilter) ([]domain.Post, int64, error) {
	return s.postRepo.List(ctx, filter)
}

// DeletePost удаляет пост
func (s *AdminService) DeletePost(ctx context.Context, actorID, postID uuid.UUID) error {
	actor, err := s.userRepo.FindByID(ctx, actorID)
	if err != nil {
		return err
	}
	if actor == nil || !actor.IsAdmin() {
		return domain.ErrInsufficientPermissions
	}
	return s.postRepo.Delete(ctx, postID)
}

// GetComments возвращает список комментариев
func (s *AdminService) GetComments(ctx context.Context, filter domain.CommentFilter) ([]domain.Comment, int64, error) {
	return s.commentRepo.List(ctx, filter)
}

// DeleteComment удаляет комментарий
func (s *AdminService) DeleteComment(ctx context.Context, actorID, commentID uuid.UUID) error {
	actor, err := s.userRepo.FindByID(ctx, actorID)
	if err != nil {
		return err
	}
	if actor == nil || !actor.IsAdmin() {
		return domain.ErrInsufficientPermissions
	}
	return s.commentRepo.Delete(ctx, commentID)
}

// ============================================
// Управление проектами и вакансиями
// ============================================

// GetProjects возвращает список проектов
func (s *AdminService) GetProjects(ctx context.Context, filter domain.ProjectFilter) ([]domain.Project, int64, error) {
	return s.projectRepo.List(ctx, filter)
}

// DeleteProject удаляет проект
func (s *AdminService) DeleteProject(ctx context.Context, actorID, projectID uuid.UUID) error {
	actor, err := s.userRepo.FindByID(ctx, actorID)
	if err != nil {
		return err
	}
	if actor == nil || !actor.IsAdmin() {
		return domain.ErrInsufficientPermissions
	}
	return s.projectRepo.Delete(ctx, projectID)
}

// GetVacancies возвращает список вакансий
func (s *AdminService) GetVacancies(ctx context.Context, filter domain.VacancyFilter) ([]domain.Vacancy, int64, error) {
	return s.vacancyRepo.List(ctx, filter)
}

// DeleteVacancy удаляет вакансию
func (s *AdminService) DeleteVacancy(ctx context.Context, actorID, vacancyID uuid.UUID) error {
	actor, err := s.userRepo.FindByID(ctx, actorID)
	if err != nil {
		return err
	}
	if actor == nil || !actor.IsAdmin() {
		return domain.ErrInsufficientPermissions
	}
	return s.vacancyRepo.Delete(ctx, vacancyID)
}

// DeleteAdmin — удаление администратора (только суперадмин)
func (s *AdminService) DeleteAdmin(ctx context.Context, actorID, adminID uuid.UUID) error {
    // Проверяем права актора (только суперадмин)
    actor, err := s.userRepo.FindByID(ctx, actorID)
    if err != nil {
        return err
    }
    if actor == nil || !actor.IsSuperAdmin() {
        return domain.ErrInsufficientPermissions
    }

    // Находим админа
    admin, err := s.userRepo.FindByID(ctx, adminID)
    if err != nil {
        return err
    }
    if admin == nil {
        return domain.ErrUserNotFound
    }

    // Нельзя удалить суперадмина
    if admin.IsSuperAdmin() {
        return domain.ErrCannotDeleteSuperAdmin
    }

    // Проверяем, что пользователь действительно админ
    if !admin.IsAdmin() {
        return domain.ErrUserNotAdmin
    }

    // Понижаем до студента
    return s.userRepo.DemoteFromAdmin(ctx, adminID)
}

// ============================================
// Статистика
// ============================================

// Stats — статистика
type Stats struct {
	TotalUsers     int64 `json:"total_users"`
	TotalAdmins    int64 `json:"total_admins"`
	TotalPosts     int64 `json:"total_posts"`
	TotalComments  int64 `json:"total_comments"`
	TotalProjects  int64 `json:"total_projects"`
	TotalVacancies int64 `json:"total_vacancies"`
	ActiveUsers    int64 `json:"active_users"`
	NewUsersToday  int64 `json:"new_users_today"`
}

// GetStats возвращает статистику
func (s *AdminService) GetStats(ctx context.Context) (*Stats, error) {
	totalUsers, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	activeUsers, err := s.userRepo.CountActive(ctx)
	if err != nil {
		return nil, err
	}

	totalAdmins, err := s.userRepo.CountAdmins(ctx)
	if err != nil {
		return nil, err
	}

	totalPosts, err := s.postRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	totalComments, err := s.commentRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	totalProjects, err := s.projectRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	totalVacancies, err := s.vacancyRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	newUsersToday, err := s.userRepo.CountNewToday(ctx)
	if err != nil {
		return nil, err
	}

	return &Stats{
		TotalUsers:     totalUsers,
		ActiveUsers:    activeUsers,
		TotalAdmins:    totalAdmins,
		TotalPosts:     totalPosts,
		TotalComments:  totalComments,
		TotalProjects:  totalProjects,
		TotalVacancies: totalVacancies,
		NewUsersToday:  newUsersToday,
	}, nil
}