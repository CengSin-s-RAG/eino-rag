package component

import (
	"agent.article.fp/config"
	"bytes"
	"encoding/json"
	"github.com/cloudwego/eino/schema"
	"io"
	"net/http"
	"strings"
)

type RewriteReq struct {
	History []string `json:"history"`
	Query   string   `json:"query"`
}

func (r *RewriteReq) GetHistory() string {
	var sb strings.Builder
	for _, s := range r.History {
		sb.WriteString(s)
		sb.WriteString("\n")
	}
	return sb.String()
}

func QueryRewriting(input []*schema.Message) (string, error) {
	var req RewriteReq
	for _, message := range input[:len(input)-1] {
		req.History = append(req.History, message.String())
	}

	req.Query = input[len(input)-1].Content
	body, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	resp, err := http.Post(config.Cfg.Rewrite.Url, "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
