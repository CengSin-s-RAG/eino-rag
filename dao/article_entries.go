package dao

import (
	"agent.article.fp/client"
	"agent.article.fp/model"
)

func GetArticlesByIds(ids []int64) ([]*model.ArticleEntries, error) {
	var contents []*model.ArticleEntries
	if err := client.IvankaContent.Model(&model.ArticleEntries{}).Where("id in ?", ids).Find(&contents).Error; err != nil {
		return nil, err
	}
	return contents, nil
}
