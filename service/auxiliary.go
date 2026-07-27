package service

import (
	"agent.article.fp/runtime"
	"agent.article.fp/store"
	"agent.article.fp/util"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type RewriteRequest struct {
	History []string `json:"history"`
	Query   string   `json:"query"`
}
type RewriteResponse struct {
	Query string `json:"query"`
}

type RerankRequest struct {
	Question  string   `json:"question"`
	Documents []string `json:"documents"`
}
type RerankItem struct {
	Index int     `json:"index"`
	ID    int64   `json:"id"`
	Score float64 `json:"score"`
}
type RerankResponse struct {
	Content []RerankItem `json:"content"`
}

type AuxiliaryService struct {
	provider runtime.Provider
	store    store.Store
}

func NewAuxiliaryService(provider runtime.Provider, store store.Store) *AuxiliaryService {
	return &AuxiliaryService{provider: provider, store: store}
}

func (s *AuxiliaryService) Rewrite(ctx context.Context, request RewriteRequest) (RewriteResponse, error) {
	if strings.TrimSpace(request.Query) == "" {
		return RewriteResponse{}, fmt.Errorf("query is required")
	}
	history := strings.Join(request.History, "\n")
	completion, err := s.provider.Complete(ctx, runtime.CompletionRequest{Messages: []runtime.Message{{Role: runtime.RoleSystem, Content: util.QueryRewritePrompt}, {Role: runtime.RoleUser, Content: fmt.Sprintf("【历史对话记录】\n%s\n【当前用户问题】\n%s", history, request.Query)}}})
	if err != nil {
		return RewriteResponse{}, err
	}
	return RewriteResponse{Query: strings.TrimSpace(completion.Message.Content)}, nil
}

func (s *AuxiliaryService) Rerank(ctx context.Context, request RerankRequest) (RerankResponse, error) {
	if strings.TrimSpace(request.Question) == "" {
		return RerankResponse{}, fmt.Errorf("question is required")
	}
	if len(request.Documents) == 0 {
		return RerankResponse{Content: []RerankItem{}}, nil
	}
	var prompt strings.Builder
	prompt.WriteString("【用户问题】\n" + request.Question + "\n【候选文档】\n")
	for index, document := range request.Documents {
		fmt.Fprintf(&prompt, "\n<document index=%d>\n%s\n</document>\n", index, document)
	}
	instruction := util.RerankPrompt + "\n仅返回 JSON，格式为 {\"content\":[{\"index\":0,\"id\":0,\"score\":0.0}]}，每个 index 必须属于候选文档，score 为 0 到 1。"
	completion, err := s.provider.Complete(ctx, runtime.CompletionRequest{Messages: []runtime.Message{{Role: runtime.RoleSystem, Content: instruction}, {Role: runtime.RoleUser, Content: prompt.String()}}})
	if err != nil {
		return RerankResponse{}, err
	}
	result, err := parseRerank(completion.Message.Content, request.Documents)
	if err != nil {
		return RerankResponse{}, err
	}
	return RerankResponse{Content: result}, nil
}

func (s *AuxiliaryService) CreateSession(ctx context.Context) (store.Session, error) {
	return s.store.CreateSession(ctx)
}
func (s *AuxiliaryService) ListSessions(ctx context.Context, offset, limit int64) ([]store.Session, error) {
	return s.store.ListSessions(ctx, offset, limit)
}
func (s *AuxiliaryService) History(ctx context.Context, id string, limit int) ([]runtime.Message, error) {
	return s.store.Recent(ctx, id, limit)
}

var idPattern = regexp.MustCompile(`【id】\s*(\d+)`)

func parseRerank(content string, documents []string) ([]RerankItem, error) {
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)
	var response RerankResponse
	if err := json.Unmarshal([]byte(content), &response); err != nil {
		return nil, fmt.Errorf("rerank model returned invalid JSON: %w", err)
	}
	seen := make(map[int]bool)
	valid := make([]RerankItem, 0, len(response.Content))
	for _, item := range response.Content {
		if item.Index < 0 || item.Index >= len(documents) || seen[item.Index] || item.Score < 0 || item.Score > 1 {
			continue
		}
		seen[item.Index] = true
		if item.ID == 0 {
			if found := idPattern.FindStringSubmatch(documents[item.Index]); len(found) == 2 {
				fmt.Sscan(found[1], &item.ID)
			}
		}
		valid = append(valid, item)
	}
	sort.Slice(valid, func(i, j int) bool { return valid[i].Score > valid[j].Score })
	return valid, nil
}
