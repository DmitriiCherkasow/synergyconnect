package domain

import "errors"

// Domain errors
var (
    // Common errors
    ErrNotFound          = errors.New("resource not found")
    ErrAlreadyExists     = errors.New("resource already exists")
    ErrInvalidInput      = errors.New("invalid input")
    ErrUnauthorized      = errors.New("unauthorized")
    ErrForbidden         = errors.New("forbidden")
    ErrInternal          = errors.New("internal server error")

    // User errors
    ErrUserNotFound      = errors.New("user not found")
    
    // Project errors
    ErrProjectNotFound   = errors.New("project not found")
    ErrProjectFull       = errors.New("project team is full")
    ErrAlreadyMember     = errors.New("user is already a member of this project")
    ErrNotMember         = errors.New("user is not a member of this project")
    ErrApplicationExists = errors.New("application already exists")
    ErrApplicationNotFound = errors.New("application not found")
    ErrCannotRemoveOwner = errors.New("cannot remove project owner")
    
    // Vacancy errors
    ErrVacancyNotFound   = errors.New("vacancy not found")
    ErrResponseExists    = errors.New("response already exists")
    ErrResponseNotFound  = errors.New("response not found")
    ErrVacancyClosed     = errors.New("vacancy is closed")
    
    // Message errors
    ErrMessageNotFound   = errors.New("message not found")
    ErrCannotSendToSelf  = errors.New("cannot send message to yourself")
    
    // 2FA errors
    ErrTwoFANotEnabled   = errors.New("2FA is not enabled")
    ErrTwoFAAlreadyEnabled = errors.New("2FA is already enabled")
    ErrInvalidCode       = errors.New("invalid 2FA code")
    ErrInvalidRecoveryCode = errors.New("invalid recovery code")
    
    // Notification errors
    ErrNotificationNotFound = errors.New("notification not found")
)