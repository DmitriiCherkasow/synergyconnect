package domain

import (
    "time"
    "github.com/google/uuid"
)

// DomainEvent represents a domain event
type DomainEvent interface {
    GetEventType() string
    GetAggregateID() uuid.UUID
    GetTimestamp() time.Time
}

// ============ Project Events ============

type ProjectCreatedEvent struct {
    ProjectID   uuid.UUID `json:"project_id"`
    OwnerID     uuid.UUID `json:"owner_id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Timestamp   time.Time `json:"timestamp"`
}

func (e ProjectCreatedEvent) GetEventType() string    { return "project.created" }
func (e ProjectCreatedEvent) GetAggregateID() uuid.UUID { return e.ProjectID }
func (e ProjectCreatedEvent) GetTimestamp() time.Time  { return e.Timestamp }

type ProjectMemberAddedEvent struct {
    ProjectID uuid.UUID `json:"project_id"`
    UserID    uuid.UUID `json:"user_id"`
    Role      string    `json:"role"`
    AddedBy   uuid.UUID `json:"added_by"`
    Timestamp time.Time `json:"timestamp"`
}

func (e ProjectMemberAddedEvent) GetEventType() string    { return "project.member_added" }
func (e ProjectMemberAddedEvent) GetAggregateID() uuid.UUID { return e.ProjectID }
func (e ProjectMemberAddedEvent) GetTimestamp() time.Time  { return e.Timestamp }

type ProjectApplicationCreatedEvent struct {
    ApplicationID uuid.UUID `json:"application_id"`
    ProjectID     uuid.UUID `json:"project_id"`
    UserID        uuid.UUID `json:"user_id"`
    Message       string    `json:"message"`
    Timestamp     time.Time `json:"timestamp"`
}

func (e ProjectApplicationCreatedEvent) GetEventType() string    { return "project.application_created" }
func (e ProjectApplicationCreatedEvent) GetAggregateID() uuid.UUID { return e.ProjectID }
func (e ProjectApplicationCreatedEvent) GetTimestamp() time.Time  { return e.Timestamp }

type ProjectApplicationStatusChangedEvent struct {
    ApplicationID uuid.UUID `json:"application_id"`
    ProjectID     uuid.UUID `json:"project_id"`
    UserID        uuid.UUID `json:"user_id"`
    OldStatus     string    `json:"old_status"`
    NewStatus     string    `json:"new_status"`
    Timestamp     time.Time `json:"timestamp"`
}

func (e ProjectApplicationStatusChangedEvent) GetEventType() string    { return "project.application_status_changed" }
func (e ProjectApplicationStatusChangedEvent) GetAggregateID() uuid.UUID { return e.ProjectID }
func (e ProjectApplicationStatusChangedEvent) GetTimestamp() time.Time  { return e.Timestamp }

// ============ Vacancy Events ============

type VacancyCreatedEvent struct {
    VacancyID  uuid.UUID `json:"vacancy_id"`
    EmployerID uuid.UUID `json:"employer_id"`
    Title      string    `json:"title"`
    Company    string    `json:"company"`
    Timestamp  time.Time `json:"timestamp"`
}

func (e VacancyCreatedEvent) GetEventType() string    { return "vacancy.created" }
func (e VacancyCreatedEvent) GetAggregateID() uuid.UUID { return e.VacancyID }
func (e VacancyCreatedEvent) GetTimestamp() time.Time  { return e.Timestamp }

type VacancyResponseCreatedEvent struct {
    ResponseID uuid.UUID `json:"response_id"`
    VacancyID  uuid.UUID `json:"vacancy_id"`
    UserID     uuid.UUID `json:"user_id"`
    Timestamp  time.Time `json:"timestamp"`
}

func (e VacancyResponseCreatedEvent) GetEventType() string    { return "vacancy.response_created" }
func (e VacancyResponseCreatedEvent) GetAggregateID() uuid.UUID { return e.VacancyID }
func (e VacancyResponseCreatedEvent) GetTimestamp() time.Time  { return e.Timestamp }

// ============ Message Events ============

type MessageCreatedEvent struct {
    MessageID  uuid.UUID `json:"message_id"`
    SenderID   uuid.UUID `json:"sender_id"`
    ReceiverID uuid.UUID `json:"receiver_id"`
    Content    string    `json:"content"`
    Timestamp  time.Time `json:"timestamp"`
}

func (e MessageCreatedEvent) GetEventType() string    { return "message.created" }
func (e MessageCreatedEvent) GetAggregateID() uuid.UUID { return e.MessageID }
func (e MessageCreatedEvent) GetTimestamp() time.Time  { return e.Timestamp }