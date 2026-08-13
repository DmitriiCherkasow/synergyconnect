package domain

import (
    "time"
    "github.com/google/uuid"
)

type VacancyStatus string

const (
    VacancyStatusActive   VacancyStatus = "active"
    VacancyStatusClosed   VacancyStatus = "closed"
    VacancyStatusArchived VacancyStatus = "archived"
)

type EmploymentType string

const (
    EmploymentTypeFullTime  EmploymentType = "full_time"
    EmploymentTypePartTime  EmploymentType = "part_time"
    EmploymentTypeContract  EmploymentType = "contract"
    EmploymentTypeInternship EmploymentType = "internship"
)

type ExperienceLevel string

const (
    ExperienceLevelEntry   ExperienceLevel = "entry"
    ExperienceLevelJunior  ExperienceLevel = "junior"
    ExperienceLevelMiddle  ExperienceLevel = "middle"
    ExperienceLevelSenior  ExperienceLevel = "senior"
    ExperienceLevelLead    ExperienceLevel = "lead"
)

type VacancyResponseStatus string

const (
    VacancyResponseStatusPending  VacancyResponseStatus = "pending"
    VacancyResponseStatusReviewed VacancyResponseStatus = "reviewed"
    VacancyResponseStatusAccepted VacancyResponseStatus = "accepted"
    VacancyResponseStatusRejected VacancyResponseStatus = "rejected"
)

type Vacancy struct {
    ID              uuid.UUID          `json:"id" gorm:"type:uuid;primary_key"`
    Title           string             `json:"title" gorm:"size:255;not null"`
    Company         string             `json:"company" gorm:"size:255;not null"`
    Description     string             `json:"description" gorm:"type:text"`
    Requirements    string             `json:"requirements" gorm:"type:text"`
    SalaryMin       *int               `json:"salary_min,omitempty"`
    SalaryMax       *int               `json:"salary_max,omitempty"`
    Currency        string             `json:"currency" gorm:"size:3;default:RUB"`
    Location        string             `json:"location" gorm:"size:255"`
    IsRemote        bool               `json:"is_remote" gorm:"default:false"`
    EmploymentType  EmploymentType     `json:"employment_type" gorm:"size:50;default:full_time"`
    ExperienceLevel ExperienceLevel    `json:"experience_level" gorm:"size:50;default:entry"`
    EmployerID      uuid.UUID          `json:"employer_id" gorm:"type:uuid;not null"`
    Employer        User               `json:"employer,omitempty" gorm:"foreignKey:EmployerID"`
    Status          VacancyStatus      `json:"status" gorm:"size:50;default:active"`
    ViewsCount      int                `json:"views_count" gorm:"default:0"`
    Responses       []VacancyResponse  `json:"responses,omitempty" gorm:"foreignKey:VacancyID"`
    CreatedAt       time.Time          `json:"created_at"`
    UpdatedAt       time.Time          `json:"updated_at"`
}

type VacancyResponse struct {
    ID          uuid.UUID              `json:"id" gorm:"type:uuid;primary_key"`
    VacancyID   uuid.UUID              `json:"vacancy_id" gorm:"type:uuid;not null"`
    Vacancy     Vacancy                `json:"vacancy,omitempty" gorm:"foreignKey:VacancyID"`
    UserID      uuid.UUID              `json:"user_id" gorm:"type:uuid;not null"`
    User        User                   `json:"user,omitempty" gorm:"foreignKey:UserID"`
    CoverLetter string                 `json:"cover_letter" gorm:"type:text"`
    Status      VacancyResponseStatus  `json:"status" gorm:"size:50;default:pending"`
    ViewedAt    *time.Time             `json:"viewed_at,omitempty"`
    CreatedAt   time.Time              `json:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at"`
}