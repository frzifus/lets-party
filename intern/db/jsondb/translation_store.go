package jsondb

import (
	"context"

	"github.com/frzifus/lets-party/intern/model"
)

func NewTranslationStore(filename string) *TranslationStore {
	return &TranslationStore{
		byLanguage: make(map[string]model.Translation),
	}
}

type TranslationStore struct {
	byLanguage map[string]model.Translation
}

func (s *TranslationStore) ListLanguages(context.Context) ([]string, error) {
	return []string{"en", "de"}, nil
}

func (s *TranslationStore) ByLanguage(_ context.Context, l string) (*model.Translation, error) {
	return &model.Translation{
		Greeting:       "Hello {{.Name}}",
		WelcomeMessage: "Welcome to the Party!",
	}, nil
}
