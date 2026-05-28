package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"logic-grpc-service/internal/config"
	"logic-grpc-service/internal/model"
	jwtutil "logic-grpc-service/internal/pkg/jwt"
	osspkg "logic-grpc-service/internal/pkg/oss"
	"logic-grpc-service/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists       = errors.New("username already exists")
	ErrInvalidCred      = errors.New("invalid username or password")
	ErrInvalidRole      = errors.New("invalid role")
	ErrForbidden        = errors.New("forbidden")
	ErrNotFound         = errors.New("not found")
	ErrProfileIncomplete = errors.New("profile incomplete")
	ErrAlreadyApplied   = errors.New("already applied")
)

type Service struct {
	repo   *repository.Repository
	cfg    *config.Config
	oss    *osspkg.Client
	aiChat *AIChatService
}

func New(repo *repository.Repository, cfg *config.Config, ossClient *osspkg.Client, aiChat *AIChatService) *Service {
	return &Service{repo: repo, cfg: cfg, oss: ossClient, aiChat: aiChat}
}

func (s *Service) Register(username, password, role string) (token string, user *model.User, err error) {
	if role != model.RoleHR && role != model.RoleCandidate {
		return "", nil, ErrInvalidRole
	}
	existing, err := s.repo.GetUserByUsername(username)
	if err != nil {
		return "", nil, err
	}
	if existing != nil {
		return "", nil, ErrUserExists
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", nil, err
	}
	user = &model.User{Username: username, PasswordHash: string(hash), Role: role}
	if err = s.repo.CreateUser(user); err != nil {
		return "", nil, err
	}
	if role == model.RoleCandidate {
		_, _ = s.repo.GetOrCreateProfile(user.ID)
	}
	token, err = jwtutil.Generate(s.cfg.JWT.Secret, s.cfg.JWT.ExpireHours, user.ID, user.Username, user.Role)
	return token, user, err
}

func (s *Service) Login(username, password string) (token string, user *model.User, err error) {
	user, err = s.repo.GetUserByUsername(username)
	if err != nil || user == nil {
		return "", nil, ErrInvalidCred
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return "", nil, ErrInvalidCred
	}
	token, err = jwtutil.Generate(s.cfg.JWT.Secret, s.cfg.JWT.ExpireHours, user.ID, user.Username, user.Role)
	return token, user, err
}

func (s *Service) GetUser(id int64) (*model.User, error) {
	return s.repo.GetUserByID(id)
}

func (s *Service) CreateJob(hrID int64, title, desc, salary, location string) (*model.Job, error) {
	job := &model.Job{
		HRID: hrID, Title: title, Description: desc,
		Salary: salary, Location: location, Status: model.JobStatusActive,
	}
	if err := s.repo.CreateJob(job); err != nil {
		return nil, err
	}
	return s.repo.GetJobByID(job.ID)
}

func (s *Service) UpdateJob(hrID, jobID int64, title, desc, salary, location, status string) (*model.Job, error) {
	job, err := s.repo.GetJobByID(jobID)
	if err != nil || job == nil {
		return nil, ErrNotFound
	}
	if job.HRID != hrID {
		return nil, ErrForbidden
	}
	if title != "" {
		job.Title = title
	}
	if desc != "" {
		job.Description = desc
	}
	if salary != "" {
		job.Salary = salary
	}
	if location != "" {
		job.Location = location
	}
	if status != "" {
		job.Status = status
	}
	if err = s.repo.UpdateJob(job); err != nil {
		return nil, err
	}
	return job, nil
}

func (s *Service) DeleteJob(hrID, jobID int64) error {
	job, err := s.repo.GetJobByID(jobID)
	if err != nil || job == nil {
		return ErrNotFound
	}
	if job.HRID != hrID {
		return ErrForbidden
	}
	return s.repo.DeleteJob(jobID)
}

func (s *Service) GetJob(jobID int64) (*model.Job, error) {
	job, err := s.repo.GetJobByID(jobID)
	if err != nil || job == nil {
		return nil, ErrNotFound
	}
	return job, nil
}

func (s *Service) ListPublicJobs(page, pageSize int) ([]model.Job, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	return s.repo.ListPublicJobs(page, pageSize, model.JobStatusActive)
}

func (s *Service) ListHRJobs(hrID int64, page, pageSize int) ([]model.Job, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	return s.repo.ListHRJobs(hrID, page, pageSize)
}

func (s *Service) GetProfile(userID int64) (*model.CandidateProfile, error) {
	return s.repo.GetOrCreateProfile(userID)
}

func (s *Service) UpdateProfile(userID int64, name, phone, education, school, experience, skills string) (*model.CandidateProfile, error) {
	profile, err := s.repo.GetOrCreateProfile(userID)
	if err != nil {
		return nil, err
	}
	profile.Name = name
	profile.Phone = phone
	profile.Education = education
	profile.School = school
	profile.Experience = experience
	profile.Skills = skills
	profile.ProfileComplete = model.IsProfileComplete(profile)
	if err = s.repo.UpdateProfile(profile); err != nil {
		return nil, err
	}
	return profile, nil
}

func (s *Service) UpdateResume(userID int64, ossKey string) (*model.CandidateProfile, error) {
	profile, err := s.repo.GetOrCreateProfile(userID)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(ossKey, fmt.Sprintf("resumes/%d/", userID)) {
		return nil, ErrForbidden
	}
	profile.ResumeOSSKey = ossKey
	profile.ProfileComplete = model.IsProfileComplete(profile)
	if err = s.repo.UpdateProfile(profile); err != nil {
		return nil, err
	}
	return profile, nil
}

func (s *Service) Apply(candidateID, jobID int64) (*model.Application, error) {
	job, err := s.repo.GetJobByID(jobID)
	if err != nil || job == nil || job.Status != model.JobStatusActive {
		return nil, ErrNotFound
	}
	profile, err := s.repo.GetProfile(candidateID)
	if err != nil || profile == nil || !model.IsProfileComplete(profile) {
		return nil, ErrProfileIncomplete
	}
	existing, err := s.repo.GetApplication(jobID, candidateID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrAlreadyApplied
	}
	app := &model.Application{JobID: jobID, CandidateID: candidateID, Status: model.AppStatusPending}
	if err = s.repo.CreateApplication(app); err != nil {
		return nil, err
	}
	return app, nil
}

func (s *Service) ListHRCandidates(hrID int64, jobID int64, page, pageSize int) ([]model.Application, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	return s.repo.ListHRCandidates(hrID, jobID, page, pageSize)
}

func (s *Service) GetUploadURL(userID int64, filename string) (string, string, int64, error) {
	return s.oss.GenerateUploadURL(userID, filename)
}

func (s *Service) GetDownloadURL(hrID int64, candidateID int64, ossKey string) (string, int64, error) {
	ok, err := s.repo.ProfileBelongsToHR(candidateID, hrID)
	if err != nil {
		return "", 0, err
	}
	if !ok {
		return "", 0, ErrForbidden
	}
	return s.oss.GenerateDownloadURL(ossKey)
}

func (s *Service) Chat(ctx context.Context, hrID int64, message string) (string, error) {
	return s.aiChat.Chat(ctx, hrID, message)
}

func (s *Service) GetChatHistory(hrID int64) ([]model.AIChatMessage, error) {
	return s.repo.ListChatMessages(hrID)
}

func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func ToJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
