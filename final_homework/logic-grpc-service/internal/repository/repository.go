package repository

import (
	"errors"

	"logic-grpc-service/internal/model"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateUser(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *Repository) GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

func (r *Repository) GetUserByID(id int64) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

func (r *Repository) CreateJob(job *model.Job) error {
	return r.db.Create(job).Error
}

func (r *Repository) GetJobByID(id int64) (*model.Job, error) {
	var job model.Job
	err := r.db.Preload("HR").First(&job, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &job, err
}

func (r *Repository) UpdateJob(job *model.Job) error {
	return r.db.Save(job).Error
}

func (r *Repository) DeleteJob(id int64) error {
	return r.db.Delete(&model.Job{}, id).Error
}

func (r *Repository) ListPublicJobs(page, pageSize int, status string) ([]model.Job, int64, error) {
	if status == "" {
		status = model.JobStatusActive
	}
	var jobs []model.Job
	var total int64
	q := r.db.Model(&model.Job{}).Where("status = ?", status)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := q.Preload("HR").Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&jobs).Error
	return jobs, total, err
}

func (r *Repository) ListHRJobs(hrID int64, page, pageSize int) ([]model.Job, int64, error) {
	var jobs []model.Job
	var total int64
	q := r.db.Model(&model.Job{}).Where("hr_id = ?", hrID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&jobs).Error
	return jobs, total, err
}

func (r *Repository) GetOrCreateProfile(userID int64) (*model.CandidateProfile, error) {
	var profile model.CandidateProfile
	err := r.db.Where("user_id = ?", userID).First(&profile).Error
	if err == nil {
		return &profile, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	profile = model.CandidateProfile{UserID: userID}
	if err := r.db.Create(&profile).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *Repository) UpdateProfile(profile *model.CandidateProfile) error {
	profile.ProfileComplete = model.IsProfileComplete(profile)
	return r.db.Save(profile).Error
}

func (r *Repository) GetProfile(userID int64) (*model.CandidateProfile, error) {
	var profile model.CandidateProfile
	err := r.db.Where("user_id = ?", userID).First(&profile).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &profile, err
}

func (r *Repository) CreateApplication(app *model.Application) error {
	return r.db.Create(app).Error
}

func (r *Repository) GetApplication(jobID, candidateID int64) (*model.Application, error) {
	var app model.Application
	err := r.db.Where("job_id = ? AND candidate_id = ?", jobID, candidateID).First(&app).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &app, err
}

func (r *Repository) ListHRCandidates(hrID int64, jobID int64, page, pageSize int) ([]model.Application, int64, error) {
	var apps []model.Application
	var total int64
	q := r.db.Model(&model.Application{}).
		Joins("JOIN jobs ON jobs.id = applications.job_id").
		Where("jobs.hr_id = ?", hrID)
	if jobID > 0 {
		q = q.Where("applications.job_id = ?", jobID)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := q.Preload("Job").Preload("Candidate").
		Order("applications.applied_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&apps).Error
	return apps, total, err
}

func (r *Repository) SaveChatMessage(msg *model.AIChatMessage) error {
	return r.db.Create(msg).Error
}

func (r *Repository) ListChatMessages(hrID int64) ([]model.AIChatMessage, error) {
	var msgs []model.AIChatMessage
	err := r.db.Where("hr_id = ?", hrID).Order("created_at ASC").Find(&msgs).Error
	return msgs, err
}

// AI stats queries

func (r *Repository) CountHRApplications(hrID int64) (int64, error) {
	var count int64
	err := r.db.Model(&model.Application{}).
		Joins("JOIN jobs ON jobs.id = applications.job_id").
		Where("jobs.hr_id = ?", hrID).
		Count(&count).Error
	return count, err
}

type JobApplicantCount struct {
	JobID    int64  `json:"job_id"`
	Title    string `json:"title"`
	Count    int64  `json:"count"`
}

func (r *Repository) JobApplicantCounts(hrID int64) ([]JobApplicantCount, error) {
	var results []JobApplicantCount
	err := r.db.Model(&model.Application{}).
		Select("jobs.id as job_id, jobs.title as title, COUNT(applications.id) as count").
		Joins("JOIN jobs ON jobs.id = applications.job_id").
		Where("jobs.hr_id = ?", hrID).
		Group("jobs.id, jobs.title").
		Order("count DESC").
		Scan(&results).Error
	return results, err
}

func (r *Repository) FilterCandidates(hrID int64, education, skill string) ([]model.CandidateProfile, error) {
	var profiles []model.CandidateProfile
	q := r.db.Model(&model.CandidateProfile{}).
		Joins("JOIN applications ON applications.candidate_id = candidate_profiles.user_id").
		Joins("JOIN jobs ON jobs.id = applications.job_id").
		Where("jobs.hr_id = ?", hrID)
	if education != "" {
		q = q.Where("candidate_profiles.education LIKE ?", "%"+education+"%")
	}
	if skill != "" {
		q = q.Where("candidate_profiles.skills LIKE ?", "%"+skill+"%")
	}
	err := q.Group("candidate_profiles.user_id").Find(&profiles).Error
	return profiles, err
}

func (r *Repository) ListHRJobsAll(hrID int64) ([]model.Job, error) {
	var jobs []model.Job
	err := r.db.Where("hr_id = ?", hrID).Order("created_at DESC").Find(&jobs).Error
	return jobs, err
}

func (r *Repository) ProfileBelongsToHR(candidateID, hrID int64) (bool, error) {
	var count int64
	err := r.db.Model(&model.Application{}).
		Joins("JOIN jobs ON jobs.id = applications.job_id").
		Where("applications.candidate_id = ? AND jobs.hr_id = ?", candidateID, hrID).
		Count(&count).Error
	return count > 0, err
}
