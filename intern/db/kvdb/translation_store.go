// Copyright (C) 2024 the lets-party maintainers
// See root-dir/LICENSE for more information

package kvdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	bolt "go.etcd.io/bbolt"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/quixsi/core/intern/model"
)

const bucketTranslation = "translation_store"

func NewTranslationStore(db *bolt.DB) (*TranslationStore, error) {
	return &TranslationStore{db: db}, db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketTranslation))
		return err
	})
}

type TranslationStore struct {
	db *bolt.DB
}

func (t *TranslationStore) UpdateLanguages(ctx context.Context, translations map[string]*model.Translation) error {
	var span trace.Span
	_, span = tracer.Start(ctx, "UpdateLanguages")
	defer span.End()

	span.AddEvent("update languages", trace.WithAttributes(attribute.Int("count", len(translations))))
	var err error
	data := make(map[string][]byte, len(translations))
	for language, translation := range translations {
		if data[language], err = json.Marshal(translation); err != nil {
			tErr := fmt.Errorf("convert translation to json: %w", err)
			span.SetStatus(codes.Error, tErr.Error())
			return tErr
		}
	}
	return t.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketTranslation))
		for lang, translation := range data {
			if err := bucket.Put([]byte(lang), translation); err != nil {
				err := fmt.Errorf("update translation for language %q", lang)
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				return err
			}
		}
		return nil
	})
}

func (t *TranslationStore) ListLanguages(ctx context.Context) ([]string, error) {
	var span trace.Span
	_, span = tracer.Start(ctx, "ListLanguages")
	defer span.End()

	span.AddEvent("View bucket")
	res := make([]string, 0)
	return res, t.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketTranslation))
		err := bucket.ForEach(func(k, _ []byte) error {
			res = append(res, string(k))
			return nil
		})
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}
		sort.Slice(res, func(i, j int) bool { return res[i] < res[j] })
		return nil
	})
}

func (t *TranslationStore) ByLanguage(ctx context.Context, l string) (*model.Translation, error) {
	var span trace.Span
	_, span = tracer.Start(ctx, "ByLanguage")
	defer span.End()
	span.AddEvent("View bucket")
	translation := &model.Translation{}
	return translation, t.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketTranslation))
		trans := bucket.Get([]byte(l))
		if trans == nil {
			err := errors.New("missing translation")
			span.RecordError(err)
			return err
		}
		return json.Unmarshal(trans, translation)
	})
}

func (t *TranslationStore) CreateLanguage(_ context.Context, key string, translation *model.Translation) error {
	return t.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketTranslation))
		val, err := json.Marshal(translation)
		if err != nil {
			return err
		}
		return bucket.Put([]byte(key), val)
	})
}
