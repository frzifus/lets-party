package jsondb

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"sync"

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

func (t *TranslationStore) ListLanguages(context.Context) ([]string, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	res := make([]string, len(t.byLanguage))
	i := 0
	for lang := range t.byLanguage {
		res[i] = lang
		i++
	}
	return res, nil
}

func (t *TranslationStore) ByLanguage(_ context.Context, l string) (*model.Translation, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	lang, ok := t.byLanguage[l]
	if !ok {
		return nil, errors.New("missing translation")
	}
	return &lang, nil
}

func (t *TranslationStore) loadFromFile() error {
	if _, err := os.Stat(t.filename); os.IsNotExist(err) {
		// File does not exist, no guests to load
		return nil
	}

	fileData, err := ioutil.ReadFile(t.filename)
	if err != nil {
		return err
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	return json.Unmarshal(fileData, &t.byLanguage)
}
