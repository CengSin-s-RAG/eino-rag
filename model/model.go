package model

import (
	"github.com/PuerkitoBio/goquery"
	"strings"
)

type ArticleEntries struct {
	Id                      int    `json:"id"`
	Title                   string `json:"title"`
	ContentShort            string `json:"content_short"`
	Content                 string `json:"content"`
	Status                  string `json:"status"`
	Precedence              int    `json:"precedence"`
	Pageviews               int    `json:"pageviews"`
	CommentDisabled         int    `json:"comment_disabled"`
	SourceUri               string `json:"source_uri"`
	SourceName              string `json:"source_name"`
	ImageUri                string `json:"image_uri"`
	Position                int    `json:"position"`
	DisplayUserId           int64  `json:"display_user_id"`
	DisplayTime             string `json:"display_time"`
	IsPriced                int    `json:"is_priced"`
	IsTrail                 int    `json:"is_trail"`
	External                int    `json:"external"`
	CreatedBy               int64  `json:"created_by"`
	UpdatedBy               int64  `json:"updated_by"`
	CreatedAt               string `json:"created_at"`
	UpdatedAt               string `json:"updated_at"`
	Platforms               string `json:"platforms"`
	Categories              string `json:"categories"`
	Tags                    string `json:"tags"`
	Symbols                 string `json:"symbols"`
	ContentPreview          string `json:"content_preview"`
	ContentParsed           string `json:"content_parsed"`
	ContentArgs             string `json:"content_args"`
	IsDeleted               int    `json:"is_deleted"`
	Extra                   string `json:"extra"`
	Score                   int    `json:"score"`
	WordsCount              int    `json:"words_count"`
	UnshowContentShort      int    `json:"unshow_content_short"`
	Themes                  string `json:"themes"`
	BaoerId                 int    `json:"baoer_id"`
	CrawlerSourceId         int    `json:"crawler_source_id"`
	Plates                  string `json:"plates"`
	IsContentCleaning       int    `json:"is_content_cleaning"`
	CrawlerWechatId         int    `json:"crawler_wechat_id"`
	ImageUris               string `json:"image_uris"`
	InfluenceScore          int    `json:"influence_score"`
	CustomTag               string `json:"custom_tag"`
	AssetTags               string `json:"asset_tags"`
	ImageScore              int    `json:"image_score"`
	Funds                   string `json:"funds"`
	Subtitle                string `json:"subtitle"`
	ShowOnWscn              int    `json:"show_on_wscn"`
	Image                   string `json:"image"`
	Images                  string `json:"images"`
	LimitedTime             int    `json:"limited_time"`
	HasTransferAudio        int    `json:"has_transfer_audio"`
	ShowLike                int    `json:"show_like"`
	PreviewParsed           string `json:"preview_parsed"`
	PreviewArgs             string `json:"preview_args"`
	ResponsibleEditorUserid string `json:"responsible_editor_userid"`
	HangDown                string `json:"hang_down"`
	References              string `json:"references"`
}

func (a *ArticleEntries) Desc() string {
	return a.ContentPreview
}

func (a *ArticleEntries) Name() string {
	return a.Title
}

func (a *ArticleEntries) ToRerankDoc() (string, error) {
	reader, err := goquery.NewDocumentFromReader(strings.NewReader(a.Content))
	if err != nil {
		return "", err
	}

	return reader.Text(), nil
}

func (a *ArticleEntries) TableName() string {
	return "article_entries"
}
