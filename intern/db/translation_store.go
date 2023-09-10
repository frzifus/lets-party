package db

import (
	"context"

	"github.com/frzifus/lets-party/intern/model"
)

type TranslationStore interface {
	ListLanguages(context.Context) ([]string, error)
	ByLanguage(context.Context, string) (*model.Translation, error)
}
