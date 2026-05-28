package handler

import (
	"net/http"
	"strconv"

	recruitv1 "logic-grpc-service/proto/gen/recruit/v1"
	"web-gin-service/internal/grpcclient"
	"web-gin-service/internal/response"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	grpc *grpcclient.Client
}

func New(grpc *grpcclient.Client) *Handler {
	return &Handler{grpc: grpc}
}

func grpcMessage(err error) string {
	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.AlreadyExists:
			return st.Message()
		case codes.InvalidArgument:
			return st.Message()
		case codes.PermissionDenied:
			return "无权限操作"
		case codes.NotFound:
			return "资源不存在"
		case codes.FailedPrecondition:
			return "请先完善个人档案并上传简历"
		default:
			return st.Message()
		}
	}
	return err.Error()
}

func grpcHTTPCode(err error) int {
	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.AlreadyExists:
			return http.StatusConflict
		case codes.InvalidArgument:
			return http.StatusBadRequest
		case codes.PermissionDenied:
			return http.StatusForbidden
		case codes.NotFound:
			return http.StatusNotFound
		case codes.FailedPrecondition:
			return http.StatusPreconditionFailed
		default:
			return http.StatusInternalServerError
		}
	}
	return http.StatusInternalServerError
}

func parsePage(c *gin.Context) *recruitv1.Pagination {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	return &recruitv1.Pagination{Page: int32(page), PageSize: int32(size)}
}

type registerReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role" binding:"required"`
}

func (h *Handler) Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	resp, err := h.grpc.Auth.Register(c, &recruitv1.RegisterRequest{
		Username: req.Username, Password: req.Password, Role: req.Role,
	})
	if err != nil {
		response.Fail(c, grpcHTTPCode(err), grpcMessage(err))
		return
	}
	response.OK(c, gin.H{"token": resp.Token, "user_id": resp.UserId, "username": resp.Username, "role": resp.Role})
}

type loginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *Handler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	resp, err := h.grpc.Auth.Login(c, &recruitv1.LoginRequest{Username: req.Username, Password: req.Password})
	if err != nil {
		response.Fail(c, grpcHTTPCode(err), grpcMessage(err))
		return
	}
	response.OK(c, gin.H{"token": resp.Token, "user_id": resp.UserId, "username": resp.Username, "role": resp.Role})
}

func (h *Handler) ListPublicJobs(c *gin.Context) {
	resp, err := h.grpc.Job.ListPublicJobs(c, &recruitv1.ListJobsRequest{Pagination: parsePage(c), Status: "active"})
	if err != nil {
		response.Fail(c, grpcHTTPCode(err), grpcMessage(err))
		return
	}
	response.OK(c, gin.H{"jobs": resp.Jobs, "page_info": resp.PageInfo})
}

func (h *Handler) GetPublicJob(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	resp, err := h.grpc.Job.GetJob(c, &recruitv1.GetJobRequest{JobId: id})
	if err != nil {
		response.Fail(c, grpcHTTPCode(err), grpcMessage(err))
		return
	}
	response.OK(c, resp.Job)
}

func (h *Handler) ListHRJobs(c *gin.Context) {
	resp, err := h.grpc.Job.ListHRJobs(c, &recruitv1.ListHRJobsRequest{
		HrId: middlewareUserID(c), Pagination: parsePage(c),
	})
	if err != nil {
		response.Fail(c, grpcHTTPCode(err), grpcMessage(err))
		return
	}
	response.OK(c, gin.H{"jobs": resp.Jobs, "page_info": resp.PageInfo})
}

type jobReq struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Salary      string `json:"salary"`
	Location    string `json:"location"`
	Status      string `json:"status"`
}

func (h *Handler) CreateJob(c *gin.Context) {
	var req jobReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	userID := middlewareUserID(c)
	resp, err := h.grpc.Job.CreateJob(c, &recruitv1.CreateJobRequest{
		HrId: userID, Title: req.Title, Description: req.Description,
		Salary: req.Salary, Location: req.Location,
	})
	if err != nil {
		response.Fail(c, grpcHTTPCode(err), grpcMessage(err))
		return
	}
	response.OK(c, resp.Job)
}

func (h *Handler) UpdateJob(c *gin.Context) {
	var req jobReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	jobID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	resp, err := h.grpc.Job.UpdateJob(c, &recruitv1.UpdateJobRequest{
		HrId: middlewareUserID(c), JobId: jobID,
		Title: req.Title, Description: req.Description,
		Salary: req.Salary, Location: req.Location, Status: req.Status,
	})
	if err != nil {
		response.Fail(c, grpcHTTPCode(err), grpcMessage(err))
		return
	}
	response.OK(c, resp.Job)
}

func (h *Handler) DeleteJob(c *gin.Context) {
	jobID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	_, err := h.grpc.Job.DeleteJob(c, &recruitv1.DeleteJobRequest{HrId: middlewareUserID(c), JobId: jobID})
	if err != nil {
		response.Fail(c, grpcHTTPCode(err), grpcMessage(err))
		return
	}
	response.OK(c, nil)
}

func (h *Handler) GetProfile(c *gin.Context) {
	resp, err := h.grpc.Candidate.GetProfile(c, &recruitv1.GetProfileRequest{UserId: middlewareUserID(c)})
	if err != nil {
		response.Fail(c, grpcHTTPCode(err), grpcMessage(err))
		return
	}
	response.OK(c, resp.Profile)
}

type profileReq struct {
	Name       string `json:"name"`
	Phone      string `json:"phone"`
	Education  string `json:"education"`
	School     string `json:"school"`
	Experience string `json:"experience"`
	Skills     string `json:"skills"`
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	var req profileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	resp, err := h.grpc.Candidate.UpdateProfile(c, &recruitv1.UpdateProfileRequest{
		UserId: middlewareUserID(c), Name: req.Name, Phone: req.Phone,
		Education: req.Education, School: req.School,
		Experience: req.Experience, Skills: req.Skills,
	})
	if err != nil {
		response.Fail(c, grpcHTTPCode(err), grpcMessage(err))
		return
	}
	response.OK(c, resp.Profile)
}

func (h *Handler) GetUploadURL(c *gin.Context) {
	filename := c.Query("filename")
	resp, err := h.grpc.OSS.GetUploadURL(c, &recruitv1.GetUploadURLRequest{
		UserId: middlewareUserID(c), Filename: filename,
	})
	if err != nil {
		response.Fail(c, grpcHTTPCode(err), grpcMessage(err))
		return
	}
	response.OK(c, gin.H{"upload_url": resp.UploadUrl, "oss_key": resp.OssKey, "expire_seconds": resp.ExpireSeconds})
}

type resumeReq struct {
	OssKey string `json:"oss_key" binding:"required"`
}

func (h *Handler) ConfirmResume(c *gin.Context) {
	var req resumeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	resp, err := h.grpc.Candidate.UpdateResume(c, &recruitv1.UpdateResumeRequest{
		UserId: middlewareUserID(c), OssKey: req.OssKey,
	})
	if err != nil {
		response.Fail(c, grpcHTTPCode(err), grpcMessage(err))
		return
	}
	response.OK(c, resp.Profile)
}

type applyReq struct {
	JobID int64 `json:"job_id" binding:"required"`
}

func (h *Handler) Apply(c *gin.Context) {
	var req applyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	resp, err := h.grpc.Application.Apply(c, &recruitv1.ApplyRequest{
		CandidateId: middlewareUserID(c), JobId: req.JobID,
	})
	if err != nil {
		response.Fail(c, grpcHTTPCode(err), grpcMessage(err))
		return
	}
	response.OK(c, resp.Application)
}

func (h *Handler) ListHRCandidates(c *gin.Context) {
	jobID, _ := strconv.ParseInt(c.Query("job_id"), 10, 64)
	resp, err := h.grpc.Application.ListHRCandidates(c, &recruitv1.ListHRCandidatesRequest{
		HrId: middlewareUserID(c), Pagination: parsePage(c), JobId: jobID,
	})
	if err != nil {
		response.Fail(c, grpcHTTPCode(err), grpcMessage(err))
		return
	}
	response.OK(c, gin.H{"applications": resp.Applications, "page_info": resp.PageInfo})
}

func (h *Handler) GetDownloadURL(c *gin.Context) {
	candidateID, _ := strconv.ParseInt(c.Query("candidate_id"), 10, 64)
	ossKey := c.Query("oss_key")
	resp, err := h.grpc.OSS.GetDownloadURL(c, &recruitv1.GetDownloadURLRequest{
		HrId: middlewareUserID(c), CandidateId: candidateID, OssKey: ossKey,
	})
	if err != nil {
		response.Fail(c, grpcHTTPCode(err), grpcMessage(err))
		return
	}
	response.OK(c, gin.H{"download_url": resp.DownloadUrl, "expire_seconds": resp.ExpireSeconds})
}

type chatReq struct {
	Message string `json:"message" binding:"required"`
}

func (h *Handler) Chat(c *gin.Context) {
	var req chatReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	resp, err := h.grpc.AI.Chat(c, &recruitv1.ChatRequest{HrId: middlewareUserID(c), Message: req.Message})
	if err != nil {
		response.Fail(c, grpcHTTPCode(err), grpcMessage(err))
		return
	}
	response.OK(c, gin.H{"reply": resp.Reply})
}

func (h *Handler) GetChatHistory(c *gin.Context) {
	resp, err := h.grpc.AI.GetChatHistory(c, &recruitv1.GetChatHistoryRequest{HrId: middlewareUserID(c)})
	if err != nil {
		response.Fail(c, grpcHTTPCode(err), grpcMessage(err))
		return
	}
	response.OK(c, gin.H{"messages": resp.Messages})
}

func middlewareUserID(c *gin.Context) int64 {
	v, ok := c.Get("user_id")
	if !ok {
		return 0
	}
	if id, ok := v.(int64); ok {
		return id
	}
	if f, ok := v.(float64); ok {
		return int64(f)
	}
	return 0
}
