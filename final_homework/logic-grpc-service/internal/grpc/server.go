package grpcserver

import (
	"context"

	recruitv1 "logic-grpc-service/proto/gen/recruit/v1"
	"logic-grpc-service/internal/model"
	"logic-grpc-service/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	recruitv1.UnimplementedAuthServiceServer
	recruitv1.UnimplementedJobServiceServer
	recruitv1.UnimplementedCandidateServiceServer
	recruitv1.UnimplementedApplicationServiceServer
	recruitv1.UnimplementedOSSServiceServer
	recruitv1.UnimplementedAIServiceServer
	svc *service.Service
}

func New(svc *service.Service) *Server {
	return &Server{svc: svc}
}

func grpcErr(err error) error {
	switch err {
	case service.ErrUserExists:
		return status.Error(codes.AlreadyExists, err.Error())
	case service.ErrInvalidCred, service.ErrInvalidRole:
		return status.Error(codes.InvalidArgument, err.Error())
	case service.ErrForbidden:
		return status.Error(codes.PermissionDenied, err.Error())
	case service.ErrNotFound:
		return status.Error(codes.NotFound, err.Error())
	case service.ErrProfileIncomplete:
		return status.Error(codes.FailedPrecondition, err.Error())
	case service.ErrAlreadyApplied:
		return status.Error(codes.AlreadyExists, "already applied")
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func toJobPB(j *model.Job) *recruitv1.Job {
	if j == nil {
		return nil
	}
	pb := &recruitv1.Job{
		Id: j.ID, HrId: j.HRID, Title: j.Title, Description: j.Description,
		Salary: j.Salary, Location: j.Location, Status: j.Status,
		CreatedAt: service.FormatTime(j.CreatedAt),
	}
	if j.HR.ID > 0 {
		pb.HrUsername = j.HR.Username
	}
	return pb
}

func toProfilePB(p *model.CandidateProfile) *recruitv1.CandidateProfile {
	if p == nil {
		return nil
	}
	return &recruitv1.CandidateProfile{
		UserId: p.UserID, Name: p.Name, Phone: p.Phone, Education: p.Education,
		School: p.School, Experience: p.Experience, Skills: p.Skills,
		ResumeOssKey: p.ResumeOSSKey, ProfileComplete: p.ProfileComplete,
	}
}

func pageInfo(page, pageSize int, total int64) *recruitv1.PageInfo {
	return &recruitv1.PageInfo{Page: int32(page), PageSize: int32(pageSize), Total: total}
}

func getPage(p *recruitv1.Pagination) (int, int) {
	page, size := 1, 10
	if p != nil {
		if p.Page > 0 {
			page = int(p.Page)
		}
		if p.PageSize > 0 {
			size = int(p.PageSize)
		}
	}
	return page, size
}

// AuthService

func (s *Server) Register(ctx context.Context, req *recruitv1.RegisterRequest) (*recruitv1.AuthResponse, error) {
	token, user, err := s.svc.Register(req.Username, req.Password, req.Role)
	if err != nil {
		return nil, grpcErr(err)
	}
	return &recruitv1.AuthResponse{Token: token, UserId: user.ID, Username: user.Username, Role: user.Role}, nil
}

func (s *Server) Login(ctx context.Context, req *recruitv1.LoginRequest) (*recruitv1.AuthResponse, error) {
	token, user, err := s.svc.Login(req.Username, req.Password)
	if err != nil {
		return nil, grpcErr(err)
	}
	return &recruitv1.AuthResponse{Token: token, UserId: user.ID, Username: user.Username, Role: user.Role}, nil
}

func (s *Server) GetUser(ctx context.Context, req *recruitv1.GetUserRequest) (*recruitv1.UserInfo, error) {
	user, err := s.svc.GetUser(req.UserId)
	if err != nil || user == nil {
		return nil, grpcErr(service.ErrNotFound)
	}
	return &recruitv1.UserInfo{Id: user.ID, Username: user.Username, Role: user.Role}, nil
}

// JobService

func (s *Server) CreateJob(ctx context.Context, req *recruitv1.CreateJobRequest) (*recruitv1.JobResponse, error) {
	job, err := s.svc.CreateJob(req.HrId, req.Title, req.Description, req.Salary, req.Location)
	if err != nil {
		return nil, grpcErr(err)
	}
	return &recruitv1.JobResponse{Job: toJobPB(job)}, nil
}

func (s *Server) UpdateJob(ctx context.Context, req *recruitv1.UpdateJobRequest) (*recruitv1.JobResponse, error) {
	job, err := s.svc.UpdateJob(req.HrId, req.JobId, req.Title, req.Description, req.Salary, req.Location, req.Status)
	if err != nil {
		return nil, grpcErr(err)
	}
	return &recruitv1.JobResponse{Job: toJobPB(job)}, nil
}

func (s *Server) DeleteJob(ctx context.Context, req *recruitv1.DeleteJobRequest) (*recruitv1.Empty, error) {
	if err := s.svc.DeleteJob(req.HrId, req.JobId); err != nil {
		return nil, grpcErr(err)
	}
	return &recruitv1.Empty{}, nil
}

func (s *Server) GetJob(ctx context.Context, req *recruitv1.GetJobRequest) (*recruitv1.JobResponse, error) {
	job, err := s.svc.GetJob(req.JobId)
	if err != nil {
		return nil, grpcErr(err)
	}
	return &recruitv1.JobResponse{Job: toJobPB(job)}, nil
}

func (s *Server) ListPublicJobs(ctx context.Context, req *recruitv1.ListJobsRequest) (*recruitv1.JobListResponse, error) {
	page, size := getPage(req.Pagination)
	jobs, total, err := s.svc.ListPublicJobs(page, size)
	if err != nil {
		return nil, grpcErr(err)
	}
	pbs := make([]*recruitv1.Job, len(jobs))
	for i := range jobs {
		pbs[i] = toJobPB(&jobs[i])
	}
	return &recruitv1.JobListResponse{Jobs: pbs, PageInfo: pageInfo(page, size, total)}, nil
}

func (s *Server) ListHRJobs(ctx context.Context, req *recruitv1.ListHRJobsRequest) (*recruitv1.JobListResponse, error) {
	page, size := getPage(req.Pagination)
	jobs, total, err := s.svc.ListHRJobs(req.HrId, page, size)
	if err != nil {
		return nil, grpcErr(err)
	}
	pbs := make([]*recruitv1.Job, len(jobs))
	for i := range jobs {
		pbs[i] = toJobPB(&jobs[i])
	}
	return &recruitv1.JobListResponse{Jobs: pbs, PageInfo: pageInfo(page, size, total)}, nil
}

// CandidateService

func (s *Server) GetProfile(ctx context.Context, req *recruitv1.GetProfileRequest) (*recruitv1.ProfileResponse, error) {
	profile, err := s.svc.GetProfile(req.UserId)
	if err != nil {
		return nil, grpcErr(err)
	}
	return &recruitv1.ProfileResponse{Profile: toProfilePB(profile)}, nil
}

func (s *Server) UpdateProfile(ctx context.Context, req *recruitv1.UpdateProfileRequest) (*recruitv1.ProfileResponse, error) {
	profile, err := s.svc.UpdateProfile(req.UserId, req.Name, req.Phone, req.Education, req.School, req.Experience, req.Skills)
	if err != nil {
		return nil, grpcErr(err)
	}
	return &recruitv1.ProfileResponse{Profile: toProfilePB(profile)}, nil
}

func (s *Server) UpdateResume(ctx context.Context, req *recruitv1.UpdateResumeRequest) (*recruitv1.ProfileResponse, error) {
	profile, err := s.svc.UpdateResume(req.UserId, req.OssKey)
	if err != nil {
		return nil, grpcErr(err)
	}
	return &recruitv1.ProfileResponse{Profile: toProfilePB(profile)}, nil
}

// ApplicationService

func (s *Server) Apply(ctx context.Context, req *recruitv1.ApplyRequest) (*recruitv1.ApplicationResponse, error) {
	app, err := s.svc.Apply(req.CandidateId, req.JobId)
	if err != nil {
		return nil, grpcErr(err)
	}
	return &recruitv1.ApplicationResponse{Application: &recruitv1.Application{
		Id: app.ID, JobId: app.JobID, CandidateId: app.CandidateID,
		Status: app.Status, AppliedAt: service.FormatTime(app.AppliedAt),
	}}, nil
}

func (s *Server) ListHRCandidates(ctx context.Context, req *recruitv1.ListHRCandidatesRequest) (*recruitv1.ApplicationListResponse, error) {
	page, size := getPage(req.Pagination)
	apps, total, err := s.svc.ListHRCandidates(req.HrId, req.JobId, page, size)
	if err != nil {
		return nil, grpcErr(err)
	}
	pbs := make([]*recruitv1.Application, len(apps))
	for i, app := range apps {
		pb := &recruitv1.Application{
			Id: app.ID, JobId: app.JobID, CandidateId: app.CandidateID,
			Status: app.Status, AppliedAt: service.FormatTime(app.AppliedAt),
		}
		if app.Job.ID > 0 {
			pb.JobTitle = app.Job.Title
		}
		if app.Candidate.ID > 0 {
			profile, _ := s.svc.GetProfile(app.CandidateID)
			pb.Candidate = toProfilePB(profile)
		}
		pbs[i] = pb
	}
	return &recruitv1.ApplicationListResponse{Applications: pbs, PageInfo: pageInfo(page, size, total)}, nil
}

// OSSService

func (s *Server) GetUploadURL(ctx context.Context, req *recruitv1.GetUploadURLRequest) (*recruitv1.GetUploadURLResponse, error) {
	url, key, expire, err := s.svc.GetUploadURL(req.UserId, req.Filename)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &recruitv1.GetUploadURLResponse{UploadUrl: url, OssKey: key, ExpireSeconds: expire}, nil
}

func (s *Server) GetDownloadURL(ctx context.Context, req *recruitv1.GetDownloadURLRequest) (*recruitv1.GetDownloadURLResponse, error) {
	url, expire, err := s.svc.GetDownloadURL(req.HrId, req.CandidateId, req.OssKey)
	if err != nil {
		return nil, grpcErr(err)
	}
	return &recruitv1.GetDownloadURLResponse{DownloadUrl: url, ExpireSeconds: expire}, nil
}

// AIService

func (s *Server) Chat(ctx context.Context, req *recruitv1.ChatRequest) (*recruitv1.ChatResponse, error) {
	reply, err := s.svc.Chat(ctx, req.HrId, req.Message)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &recruitv1.ChatResponse{Reply: reply}, nil
}

func (s *Server) GetChatHistory(ctx context.Context, req *recruitv1.GetChatHistoryRequest) (*recruitv1.ChatHistoryResponse, error) {
	msgs, err := s.svc.GetChatHistory(req.HrId)
	if err != nil {
		return nil, grpcErr(err)
	}
	pbs := make([]*recruitv1.ChatMessage, len(msgs))
	for i, m := range msgs {
		pbs[i] = &recruitv1.ChatMessage{
			Id: m.ID, Role: m.Role, Content: m.Content,
			CreatedAt: service.FormatTime(m.CreatedAt),
		}
	}
	return &recruitv1.ChatHistoryResponse{Messages: pbs}, nil
}
