package bridge

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Ja7ad/meilibridge/pkg/types"
	meili "github.com/meilisearch/meilisearch-go"
	"io"
	"net/http"
	"strings"

	"github.com/Ja7ad/meilibridge/config"
	"github.com/Ja7ad/meilibridge/pkg/database"
	"github.com/Ja7ad/meilibridge/pkg/meilisearch"
)

func updateItemKeys(results []*database.Result, fields map[string]string) {
	if fields == nil {
		return
	}

	for i := range results {
		resultMap := *results[i]

		for key := range resultMap {
			if _, exists := fields[key]; !exists {
				delete(resultMap, key)
			}
		}

		for fk, fv := range fields {
			if fv != "" {
				if value, exists := resultMap[fk]; exists {
					resultMap[fv] = value
					delete(resultMap, fk)
				}
			}
		}
	}
}

func recreateIndex(
	ctx context.Context,
	indexName string,
	primaryKey string,
	set *config.Settings,
	meili meilisearch.Meilisearch,
) error {
	if meili.IsExistsIndex(indexName) {
		if err := meili.DeleteIndex(ctx, indexName); err != nil {
			return err
		}
	}

	if err := meili.CreateIndex(ctx, indexName, primaryKey); err != nil {
		return err
	}

	if set != nil {
		if err := meili.UpdateIndexSettings(ctx, indexName, set); err != nil {
			return err
		}
	}

	return nil
}

func progressBar(totalItems, totalIndexedItems int64, bridge, col, index string) {
	percentage := float64(totalIndexedItems) / float64(totalItems) * 100
	barLength := 50
	filledLength := int(float64(barLength) * percentage / 100)
	bar := strings.Repeat("=", filledLength) + strings.Repeat(" ", barLength-filledLength)
	fmt.Printf("\r%.0f%% [%s] %s - %s -> %s (%d/%d)\n",
		percentage, bar,
		bridge,
		col, index,
		totalIndexedItems, totalItems,
	)
}

func isValidTriggerToken(r *http.Request, token string) bool {
	if len(token) == 0 {
		return true
	}

	return r.Header.Get(_triggerHeaderKey) == token
}

func unmarshalTriggerBody(body io.ReadCloser) (*types.TriggerRequestBody, error) {
	r := new(types.TriggerRequestBody)
	return r, json.NewDecoder(body).Decode(r)
}

func triggerHandler(token string, queue *Queue) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if !isValidTriggerToken(r, token) {
			http.Error(w, "invalid trigger token", http.StatusUnauthorized)
			return
		}

		b, err := unmarshalTriggerBody(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		if err := b.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		queue.Add(*b)

		w.WriteHeader(http.StatusAccepted)
	}
}

func indexConfigByUID(uid string, indexMap map[config.Collection]*config.IndexConfig) (string, *config.IndexConfig) {
	for col, idx := range indexMap {
		if uid == idx.IndexName {
			return col.GetView(), idx
		}
	}
	return "", nil
}

func processTrigger(
	ctx context.Context,
	waitFunc func(ctx context.Context, t *meili.TaskInfo) error,
	idx *meili.Index,
	triggerType types.TriggerOpType,
	document database.Result,
	primaryValue string,
) error {

	switch triggerType {
	case types.INSERT, types.UPDATE:
		task, err := idx.UpdateDocuments(document)
		if err != nil {
			return err
		}

		if err := waitFunc(ctx, task); err != nil {
			return err
		}
	case types.DELETE:
		task, err := idx.DeleteDocument(primaryValue)
		if err != nil {
			return err
		}

		if err := waitFunc(ctx, task); err != nil {
			return err
		}
	}
	return nil
}
