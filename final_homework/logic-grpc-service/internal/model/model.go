package model

import "time"

const (
	RoleHR        = "hr"
	RoleCandidate = "candidate"

	JobStatusActive   = "active"
	JobStatusInactive = "inactive"

	AppStatusPending = "pending"
)

type User struct {
	ID           int64     `gorm:"primaryKey;autoIncrement"`
	Username     string    `gorm:"size:64;uniqueIndex;not null"`
	PasswordHash string    `gorm:"size:255;not null"`
	Role         string    `gorm:"size:16;not null;index"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}

func (User) TableName() string { return "users" }

type Job struct {
	ID          int64     `gorm:"primaryKey;autoIncrement"`
	HRID        int64     `gorm:"not null;index"`
	Title       string    `gorm:"size:128;not null"`
	Description string    `gorm:"type:text"`
	Salary      string    `gorm:"size:64"`
	Location    string    `gorm:"size:128"`
	Status      string    `gorm:"size:16;default:active;index"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	HR          User      `gorm:"foreignKey:HRID"`
}

func (Job) TableName() string { return "jobs" }

type CandidateProfile struct {
	UserID          int64     `gorm:"primaryKey"`
	Name            string    `gorm:"size:64"`
	Phone           string    `gorm:"size:32"`
	Education       string    `gorm:"size:64"`
	School          string    `gorm:"size:128"`
	Experience      string    `gorm:"type:text"`
	Skills          string    `gorm:"type:text"`
	ResumeOSSKey    string    `gorm:"size:512"`
	ProfileComplete bool      `gorm:"default:false"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime"`
	User            User      `gorm:"foreignKey:UserID"`
}

func (CandidateProfile) TableName() string { return "candidate_profiles" }

type Application struct {
	ID          int64     `gorm:"primaryKey;autoIncrement"`
	JobID       int64     `gorm:"not null;index"`
	CandidateID int64     `gorm:"not null;index"`
	Status      string    `gorm:"size:16;default:pending"`
	AppliedAt   time.Time `gorm:"autoCreateTime"`
	Job         Job       `gorm:"foreignKey:JobID"`
	Candidate   User      `gorm:"foreignKey:CandidateID"`
}

func (Application) TableName() string { return "applications" }

type AIChatMessage struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	HRID      int64     `gorm:"not null;index"`
	Role      string    `gorm:"size:16;not null"`
	Content   string    `gorm:"type:text;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime;index"`
}

func (AIChatMessage) TableName() string { return "ai_chat_messages" }

func IsProfileComplete(p *CandidateProfile) bool {
	return p.Name != "" && p.Phone != "" && p.Education != "" &&
		p.School != "" && p.Experience != "" && p.Skills != "" &&
		p.ResumeOSSKey != ""
}
