package bridge

import (
	"context"
	"github.com/Ja7ad/meilibridge/pkg/internal/types"
	"github.com/Ja7ad/meilibridge/pkg/logger"
	"time"
)

type Queue struct {
	items chan types.TriggerRequestBody
	log   logger.Logger
}

func newQueue(log logger.Logger) *Queue {
	return &Queue{
		items: make(chan types.TriggerRequestBody),
		log:   log,
	}
}

func (q *Queue) Add(item types.TriggerRequestBody) {
	q.items <- item
	q.log.Info("add new item to queue",
		"index", item.IndexUID,
		"operation", item.Type,
		"document", item.Document,
	)
}

func (q *Queue) Process(ctx context.Context, processFunc func(ctx context.Context, i types.TriggerRequestBody) (bool, error)) {
	for {
		select {
		case <-ctx.Done():
			q.log.Info("stopping queue")
			close(q.items)
			return
		case item := <-q.items:
			requeue, err := processFunc(ctx, item)
			if err != nil {
				q.log.Error("failed to process item, requeue it after 5 second",
					"index", item.IndexUID,
					"operation", item.Type,
					"document", item.Document,
					"error", err,
				)
				if requeue {
					go func(i types.TriggerRequestBody) {
						time.Sleep(5 * time.Second)
						q.Add(i)
					}(item)
				}
			} else {
				q.log.Info("processed item", "index", item.IndexUID, "operation", item.Type)
			}
		}
	}
}
