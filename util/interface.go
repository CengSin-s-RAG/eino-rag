package util

type RerankDoc interface {
	ToRerankDoc() (string, error)
	Desc() string
	Name() string
}
