// Copyright (C) 2024 the lets-party maintainers
// See root-dir/LICENSE for more information

package jsondb

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sort"
	"sync"

	"go.opentelemetry.io/otel/trace"

	"github.com/frzifus/lets-party/intern/model"
)

func NewTranslationStore(filename string) (*TranslationStore, error) {
	store := &TranslationStore{
		filename:   filename,
		byLanguage: make(map[string]model.Translation),
	}
	if err := store.loadFromFile(); err != nil {
		return nil, err
	}
	return store, nil
}

type TranslationStore struct {
	mu sync.RWMutex

	filename   string
	byLanguage map[string]model.Translation
}

func (t *TranslationStore) ListLanguages(ctx context.Context) ([]string, error) {
	var span trace.Span
	_, span = tracer.Start(ctx, "ListLanguages")
	defer span.End()

	span.AddEvent("Lock")
	t.mu.Lock()
	defer span.AddEvent("RUlock")
	defer t.mu.Unlock()

	res := make([]string, len(t.byLanguage))
	i := 0
	for lang := range t.byLanguage {
		res[i] = lang
		i++
	}
	sort.Slice(res, func(i, j int) bool { return res[i] < res[j] })
	return res, nil
}

func (t *TranslationStore) ByLanguage(ctx context.Context, l string) (*model.Translation, error) {
	var span trace.Span
	_, span = tracer.Start(ctx, "ByLanguage")
	defer span.End()

	span.AddEvent("RLock")
	t.mu.RLock()
	defer span.AddEvent("RUnlock")
	defer t.mu.RUnlock()

	lang, ok := t.byLanguage[l]
	if !ok {
		err := errors.New("missing translation")
		span.RecordError(err)
		return nil, err
	}
	return &lang, nil
}

func (t *TranslationStore) loadFromFile() error {
	if _, err := os.Stat(t.filename); os.IsNotExist(err) {
		// File does not exist, no guests to load
		return nil
	}

	fileData, err := os.ReadFile(t.filename)
	if err != nil {
		return err
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	return json.Unmarshal(fileData, &t.byLanguage)
}

func (t *TranslationStore) CreateLanguage(context.Context, string, *model.Translation) error {
	return errors.New("not implemented")
}

func (t *TranslationStore) UpdateLanguages(context.Context, map[string]*model.Translation) error {
	return errors.New("not implemented")
}
