package db

import (
	"context"

	"github.com/frzifus/lets-party/intern/model"
)

type TranslationStore interface {
	ListLanguages(context.Context) ([]string, error)
	ByLanguage(context.Context, string) (*model.Translation, error)
	CreateLanguage(context.Context, string, *model.Translation) error
	UpdateLanguages(context.Context, map[string]*model.Translation) error
}
