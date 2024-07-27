package bridge

import (
	"context"
	"fmt"
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

func progressBar(totalItems, totalIndexedItems int64, col, index string) {
	percentage := float64(totalIndexedItems) / float64(totalItems) * 100
	barLength := 50
	filledLength := int(float64(barLength) * percentage / 100)
	bar := strings.Repeat("=", filledLength) + strings.Repeat(" ", barLength-filledLength)
	fmt.Printf("\r%.0f%% [%s] (%d/%d) %s -> %s\n",
		percentage, bar,
		totalIndexedItems, totalItems,
		col, index,
	)
}
