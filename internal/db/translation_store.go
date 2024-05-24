// Copyright (C) 2024 the quixsi maintainers
// See root-dir/LICENSE for more information

package db

import (
	"context"

	"github.com/quixsi/core/internal/model"
)

type TranslationStore interface {
	ListLanguages(context.Context) ([]string, error)
	ByLanguage(context.Context, string) (*model.Translation, error)
	CreateLanguage(context.Context, string, *model.Translation) error
	UpdateLanguages(context.Context, map[string]*model.Translation) error
}
