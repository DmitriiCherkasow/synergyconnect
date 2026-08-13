package domain

import (
    "time"
    "github.com/google/uuid"
)

type ProjectStatus string

const (
    ProjectStatusOpen       ProjectStatus = "open"
    ProjectStatusInProgress ProjectStatus = "in_progress"
    ProjectStatusCompleted  ProjectStatus = "completed"
    ProjectStatusArchived   ProjectStatus = "archived"
)

type ProjectMemberRole string

const (
    ProjectMemberRoleOwner  ProjectMemberRole = "owner"
    ProjectMemberRoleAdmin  ProjectMemberRole = "admin"
    ProjectMemberRoleMember ProjectMemberRole = "member"
)

type ProjectApplicationStatus string

const (
    ProjectApplicationStatusPending  ProjectApplicationStatus = "pending"
    ProjectApplicationStatusAccepted ProjectApplicationStatus = "accepted"
    ProjectApplicationStatusRejected ProjectApplicationStatus = "rejected"
)

type Project struct {
    ID          uuid.UUID          `json:"id" gorm:"type:uuid;primary_key"`
    Title       string             `json:"title" gorm:"size:255;not null"`
    Description string             `json:"description" gorm:"type:text"`
    Status      ProjectStatus      `json:"status" gorm:"size:50;default:open"`
    OwnerID     uuid.UUID          `json:"owner_id" gorm:"type:uuid;not null"`
    Owner       User               `json:"owner,omitempty" gorm:"foreignKey:OwnerID"`
    MaxTeamSize int                `json:"max_team_size" gorm:"default:5"`
    IsPublic    bool               `json:"is_public" gorm:"default:true"`
    Tags        []string           `json:"tags" gorm:"type:text[]"`
    Members     []ProjectMember    `json:"members,omitempty" gorm:"foreignKey:ProjectID"`
    Applications []ProjectApplication `json:"applications,omitempty" gorm:"foreignKey:ProjectID"`
    CreatedAt   time.Time          `json:"created_at"`
    UpdatedAt   time.Time          `json:"updated_at"`
}

type ProjectMember struct {
    ProjectID uuid.UUID          `json:"project_id" gorm:"type:uuid;primaryKey"`
    UserID    uuid.UUID          `json:"user_id" gorm:"type:uuid;primaryKey"`
    User      User               `json:"user,omitempty" gorm:"foreignKey:UserID"`
    Role      ProjectMemberRole  `json:"role" gorm:"size:50;default:member"`
    JoinedAt  time.Time          `json:"joined_at"`
}

type ProjectApplication struct {
    ID        uuid.UUID                 `json:"id" gorm:"type:uuid;primary_key"`
    ProjectID uuid.UUID                 `json:"project_id" gorm:"type:uuid;not null"`
    Project   Project                   `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
    UserID    uuid.UUID                 `json:"user_id" gorm:"type:uuid;not null"`
    User      User                      `json:"user,omitempty" gorm:"foreignKey:UserID"`
    Message   string                    `json:"message" gorm:"type:text"`
    Status    ProjectApplicationStatus  `json:"status" gorm:"size:50;default:pending"`
    CreatedAt time.Time                 `json:"created_at"`
    UpdatedAt time.Time                 `json:"updated_at"`
}

// Domain методы
func (p *Project) IsMember(userID uuid.UUID) bool {
    for _, member := range p.Members {
        if member.UserID == userID {
            return true
        }
    }
    return false
}

func (p *Project) IsOwner(userID uuid.UUID) bool {
    return p.OwnerID == userID
}

func (p *Project) CanManage(userID uuid.UUID) bool {
    if p.IsOwner(userID) {
        return true
    }
    for _, member := range p.Members {
        if member.UserID == userID && (member.Role == ProjectMemberRoleAdmin || member.Role == ProjectMemberRoleOwner) {
            return true
        }
    }
    return false
}

func (p *Project) HasPendingApplication(userID uuid.UUID) bool {
    for _, app := range p.Applications {
        if app.UserID == userID && app.Status == ProjectApplicationStatusPending {
            return true
        }
    }
    return false
}