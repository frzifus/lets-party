package kvdb

import (
	"context"
	"encoding/json"
	"errors"
	"sort"

	bolt "go.etcd.io/bbolt"
	"go.opentelemetry.io/otel/trace"

	"github.com/frzifus/lets-party/intern/model"
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

func (t *TranslationStore) ListLanguages(ctx context.Context) ([]string, error) {
	var span trace.Span
	ctx, span = tracer.Start(ctx, "ListLanguages")
	defer span.End()

	span.AddEvent("View bucket")
	res := make([]string, 0)
	return res, t.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketTranslation))
		bucket.ForEach(func(k, _ []byte) error {
			res = append(res, string(k))
			return nil
		})
		sort.Slice(res, func(i, j int) bool { return res[i] < res[j] })
		return nil
	})
}

func (t *TranslationStore) ByLanguage(ctx context.Context, l string) (*model.Translation, error) {
	var span trace.Span
	ctx, span = tracer.Start(ctx, "ByLanguage")
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
