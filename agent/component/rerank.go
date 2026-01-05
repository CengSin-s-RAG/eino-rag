package component

import (
	"agent.article.fp/config"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sort"
	"strings"
)

type RerankState struct {
	Items []*Resp `json:"content"`
}

// {
//  "role" : "assistant",
//  "content" : "{\n  \"id\": 3757629,\n  \"index\": 2,\n  \"score\": 0.95\n}",
//  "response_meta" : {
//    "finish_reason" : "stop",
//    "usage" : {
//      "prompt_tokens" : 11667,
//      "prompt_token_details" : {
//        "cached_tokens" : 0
//      },
//      "completion_tokens" : 32,
//      "total_tokens" : 11699,
//      "completion_token_details" : { }
//    }
//  },
//  "extra" : {
//    "openai-request-id" : "gen-1767168874-5JTkCdaKtO3JpVVqa7ti"
//  }
//}

type Resp struct {
	Index int64   `json:"index" jsonschema:"title=Index,descrption=文档的下标"`
	Id    int     `json:"id" jsonschema:"title=id,description=参考文档的id"`
	Score float64 `json:"score" jsonschema:"title=score,description=参考文档相对于用户问题的得分"`
}

type RerankReq struct {
	Question  string   `json:"question"`
	Documents []string `json:"documents"`
}

func Rerank(ctx context.Context, body *RerankReq) ([]*Resp, error) {
	if body == nil {
		return nil, errors.New("body is nil")
	}
	bytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(config.Cfg.Rerank.Url, "application/json; charset=utf-8", strings.NewReader(string(bytes)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	all, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rsp RerankState
	if err = json.Unmarshal(all, &rsp); err != nil {
		return nil, err
	}

	sort.Slice(rsp.Items, func(i, j int) bool {
		return rsp.Items[i].Score > rsp.Items[j].Score
	})
	return rsp.Items, nil
}
