package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"logic-grpc-service/internal/config"
	"logic-grpc-service/internal/model"
	aipkg "logic-grpc-service/internal/pkg/ai"
	"logic-grpc-service/internal/repository"

	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

type AIChatService struct {
	repo      *repository.Repository
	chat      einomodel.ChatModel
	modelName string
}

func NewAIChatService(repo *repository.Repository, cfg config.DashScopeConfig) (*AIChatService, error) {
	chatModel := aipkg.NewDashScopeChatModel(aipkg.DashScopeConfig{
		APIKey:  cfg.APIKey,
		BaseURL: cfg.BaseURL,
		Model:   cfg.Model,
	})
	return &AIChatService{repo: repo, chat: chatModel, modelName: cfg.Model}, nil
}

type intentType string

const (
	intentTotalApplicants  intentType = "total_applicants"
	intentJobApplicantCount intentType = "job_applicant_count"
	intentPopularJobs      intentType = "popular_jobs"
	intentFilterCandidates intentType = "filter_candidates"
	intentMyJobsSummary    intentType = "my_jobs_summary"
	intentGeneral          intentType = "general"
)

type intentResult struct {
	Type      intentType `json:"type"`
	Education string     `json:"education,omitempty"`
	Skill     string     `json:"skill,omitempty"`
}

func parseIntent(message string) intentResult {
	msg := strings.ToLower(message)
	switch {
	case strings.Contains(msg, "总") && (strings.Contains(msg, "投递") || strings.Contains(msg, "申请") || strings.Contains(msg, "候选人")):
		return intentResult{Type: intentTotalApplicants}
	case strings.Contains(msg, "热门") || strings.Contains(msg, "排行") || strings.Contains(msg, "最多"):
		return intentResult{Type: intentPopularJobs}
	case strings.Contains(msg, "岗位") && (strings.Contains(msg, "多少") || strings.Contains(msg, "投递") || strings.Contains(msg, "申请")):
		return intentResult{Type: intentJobApplicantCount}
	case strings.Contains(msg, "筛选") || strings.Contains(msg, "过滤") || strings.Contains(msg, "查找"):
		ir := intentResult{Type: intentFilterCandidates}
		for _, edu := range []string{"博士", "硕士", "本科", "大专", "高中"} {
			if strings.Contains(message, edu) {
				ir.Education = edu
				break
			}
		}
		if strings.Contains(msg, "技能") {
			parts := strings.Split(message, "技能")
			if len(parts) > 1 {
				ir.Skill = strings.TrimSpace(parts[1])
			}
		}
		return ir
	case strings.Contains(msg, "我的岗位") || strings.Contains(msg, "发布") && strings.Contains(msg, "岗位"):
		return intentResult{Type: intentMyJobsSummary}
	default:
		return intentResult{Type: intentGeneral}
	}
}

func (a *AIChatService) buildContext(hrID int64, message string) (string, error) {
	intent := parseIntent(message)
	var data interface{}
	var err error

	switch intent.Type {
	case intentTotalApplicants:
		count, e := a.repo.CountHRApplications(hrID)
		if e != nil {
			return "", e
		}
		data = map[string]interface{}{"total_applicants": count}
	case intentJobApplicantCount, intentPopularJobs:
		counts, e := a.repo.JobApplicantCounts(hrID)
		if e != nil {
			return "", e
		}
		data = counts
	case intentFilterCandidates:
		profiles, e := a.repo.FilterCandidates(hrID, intent.Education, intent.Skill)
		if e != nil {
			return "", e
		}
		data = profiles
	case intentMyJobsSummary:
		jobs, e := a.repo.ListHRJobsAll(hrID)
		if e != nil {
			return "", e
		}
		counts, _ := a.repo.JobApplicantCounts(hrID)
		data = map[string]interface{}{"jobs": jobs, "applicant_counts": counts}
	default:
		count, _ := a.repo.CountHRApplications(hrID)
		counts, _ := a.repo.JobApplicantCounts(hrID)
		jobs, _ := a.repo.ListHRJobsAll(hrID)
		data = map[string]interface{}{
			"total_applicants": count,
			"job_applicant_counts": counts,
			"my_jobs": jobs,
		}
	}

	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (a *AIChatService) Chat(ctx context.Context, hrID int64, message string) (string, error) {
	_ = a.repo.SaveChatMessage(&model.AIChatMessage{HRID: hrID, Role: "user", Content: message})

	contextData, err := a.buildContext(hrID, message)
	if err != nil {
		return "", err
	}

	systemPrompt := `你是智能招聘系统的 HR 数据分析助手。请根据提供的 MySQL 业务统计数据，用简洁专业的中文回答 HR 的问题。
要求：
1. 只基于提供的业务数据回答，不要编造数据
2. 如果数据为空，如实说明
3. 回答要结构清晰，必要时使用列表`

	userPrompt := fmt.Sprintf("业务统计数据（JSON）：\n%s\n\nHR 问题：%s", contextData, message)

	resp, err := a.chat.Generate(ctx, []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(userPrompt),
	})
	if err != nil {
		return "", fmt.Errorf("eino generate: %w", err)
	}

	reply := resp.Content
	_ = a.repo.SaveChatMessage(&model.AIChatMessage{HRID: hrID, Role: "assistant", Content: reply})
	return reply, nil
}
