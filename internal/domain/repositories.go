package domain

import (
    "context"

    "github.com/google/uuid"
)

// ProjectRepository defines project repository interface
type ProjectRepository interface {
    Create(ctx context.Context, project *Project) error
    GetByID(ctx context.Context, id uuid.UUID) (*Project, error)
    Update(ctx context.Context, project *Project) error
    Delete(ctx context.Context, id uuid.UUID) error
    List(ctx context.Context, filter ProjectFilter) ([]Project, int64, error)
    AddMember(ctx context.Context, member *ProjectMember) error
    RemoveMember(ctx context.Context, projectID, userID uuid.UUID) error
    GetMembers(ctx context.Context, projectID uuid.UUID) ([]ProjectMember, error)
    CreateApplication(ctx context.Context, application *ProjectApplication) error
    UpdateApplication(ctx context.Context, application *ProjectApplication) error
    GetApplication(ctx context.Context, id uuid.UUID) (*ProjectApplication, error)
    GetUserApplications(ctx context.Context, userID uuid.UUID) ([]ProjectApplication, error)
    Count(ctx context.Context) (int64, error)
}

type ProjectFilter struct {
    Status   *ProjectStatus
    OwnerID  *uuid.UUID
    MemberID *uuid.UUID
    Tag      string
    Search   string
    Limit    int
    Offset   int
}

// VacancyRepository defines vacancy repository interface
type VacancyRepository interface {
    Create(ctx context.Context, vacancy *Vacancy) error
    GetByID(ctx context.Context, id uuid.UUID) (*Vacancy, error)
    Update(ctx context.Context, vacancy *Vacancy) error
    Delete(ctx context.Context, id uuid.UUID) error
    List(ctx context.Context, filter VacancyFilter) ([]Vacancy, int64, error)
    IncrementViews(ctx context.Context, id uuid.UUID) error
    CreateResponse(ctx context.Context, response *VacancyResponse) error
    UpdateResponse(ctx context.Context, response *VacancyResponse) error
    GetResponse(ctx context.Context, id uuid.UUID) (*VacancyResponse, error)
    GetResponsesByVacancy(ctx context.Context, vacancyID uuid.UUID) ([]VacancyResponse, error)
    GetUserResponses(ctx context.Context, userID uuid.UUID) ([]VacancyResponse, error)
    Search(ctx context.Context, query string, limit, offset int) ([]Vacancy, int64, error)
    Count(ctx context.Context) (int64, error)
}

type VacancyFilter struct {
    Company        *string
    Location       *string
    IsRemote       *bool
    EmploymentType *EmploymentType
    ExperienceLevel *ExperienceLevel
    SalaryMin      *int
    SalaryMax      *int
    Status         *VacancyStatus
    Search         string
    Limit          int
    Offset         int
}

// MessageRepository defines message repository interface
type MessageRepository interface {
    Create(ctx context.Context, message *Message) error
    GetByID(ctx context.Context, id uuid.UUID) (*Message, error)
    GetConversation(ctx context.Context, user1ID, user2ID uuid.UUID, limit, offset int) ([]Message, int64, error)
    GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error)
    GetUnreadCountFromUser(ctx context.Context, userID, senderID uuid.UUID) (int64, error)
    MarkAsRead(ctx context.Context, messageID uuid.UUID) error
    MarkAllAsRead(ctx context.Context, userID, senderID uuid.UUID) error
    GetRecentChats(ctx context.Context, userID uuid.UUID, limit int) ([]Message, error)
    Delete(ctx context.Context, id uuid.UUID) error
    DeleteConversation(ctx context.Context, user1ID, user2ID uuid.UUID) error
}

// NotificationRepository defines notification repository interface
type NotificationRepository interface {
    Create(ctx context.Context, notification *Notification) error
    GetByID(ctx context.Context, id uuid.UUID) (*Notification, error)
    GetByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Notification, int64, error)
    GetUnreadByUser(ctx context.Context, userID uuid.UUID) ([]Notification, error)
    GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error)
    MarkAsRead(ctx context.Context, id uuid.UUID) error
    MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
    Delete(ctx context.Context, id uuid.UUID) error
    DeleteAllByUser(ctx context.Context, userID uuid.UUID) error
}

// TwoFARepository defines 2FA repository interface
type TwoFARepository interface {
    CreateOrUpdate(ctx context.Context, twofa *UserTwoFA) error
    GetByUserID(ctx context.Context, userID uuid.UUID) (*UserTwoFA, error)
    Enable(ctx context.Context, userID uuid.UUID, secret string) error
    Disable(ctx context.Context, userID uuid.UUID) error
    IsEnabled(ctx context.Context, userID uuid.UUID) (bool, error)
    UpdateRecoveryCodes(ctx context.Context, userID uuid.UUID, codes []string) error
    Delete(ctx context.Context, userID uuid.UUID) error
}

// DeviceRepository defines device repository interface
type DeviceRepository interface {
    Create(ctx context.Context, device *UserDevice) error
    GetByUserID(ctx context.Context, userID uuid.UUID) ([]UserDevice, error)
    UpdateLastUsed(ctx context.Context, id uuid.UUID) error
    Delete(ctx context.Context, id uuid.UUID) error
}

// UserRepository defines user repository interface
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	Update(ctx context.Context, user *User) error
	UpdateLastLogin(ctx context.Context, userID uuid.UUID) error
    List(ctx context.Context, filter UserFilter) ([]User, int64, error)
    Delete(ctx context.Context, id uuid.UUID) error

    // Admin specific methods
    FindSuperAdmin(ctx context.Context) (*User, error)
    CountAdmins(ctx context.Context) (int64, error)
    FindAdmins(ctx context.Context) ([]User, error)
    PromoteToAdmin(ctx context.Context, userID uuid.UUID) error
    DemoteFromAdmin(ctx context.Context, userID uuid.UUID) error

    // Stats methods
    Count(ctx context.Context) (int64, error)
    CountActive(ctx context.Context) (int64, error)
    CountNewToday(ctx context.Context) (int64, error)
}

// UserFilter — фильтр для списка пользователей
type UserFilter struct {
    Search   string
    Role     *UserRole
    IsActive *bool
    Limit    int
    Offset   int
}

// PostRepository defines post repository interface
type PostRepository interface {
    Create(ctx context.Context, post *Post) error
    GetByID(ctx context.Context, id uuid.UUID) (*Post, error)
    Update(ctx context.Context, post *Post) error
    Delete(ctx context.Context, id uuid.UUID) error
    List(ctx context.Context, filter PostFilter) ([]Post, int64, error)
    ListByAuthor(ctx context.Context, authorID uuid.UUID, limit, offset int) ([]Post, int64, error)
    ListByGroup(ctx context.Context, groupID uuid.UUID, limit, offset int) ([]Post, int64, error)
    GetPublicFeed(ctx context.Context, limit, offset int) ([]Post, int64, error)
    GetFeed(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Post, int64, error)
    Count(ctx context.Context) (int64, error)
}

// CommentRepository defines comment repository interface
type CommentRepository interface {
    Create(ctx context.Context, comment *Comment) error
    GetByID(ctx context.Context, id uuid.UUID) (*Comment, error)
    Update(ctx context.Context, comment *Comment) error
    Delete(ctx context.Context, id uuid.UUID) error
    ListByPost(ctx context.Context, postID uuid.UUID) ([]Comment, error)
    ListByAuthor(ctx context.Context, authorID uuid.UUID, limit, offset int) ([]Comment, int64, error)
    List(ctx context.Context, filter CommentFilter) ([]Comment, int64, error)
    Count(ctx context.Context) (int64, error)
    CountByPostID(ctx context.Context, postID uuid.UUID) (int64, error)
}

// PostFilter — фильтр для постов
type PostFilter struct {
    AuthorID   *uuid.UUID
    GroupID    *uuid.UUID
    Category   *PostCategory
    Visibility *PostVisibility
    Search     string
    Limit      int
    Offset     int
}

// CommentFilter — фильтр для комментариев
type CommentFilter struct {
    PostID     *uuid.UUID
    AuthorID   *uuid.UUID
    Search     string
    Limit      int
    Offset     int
}