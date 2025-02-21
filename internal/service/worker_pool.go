package service

import (
	"context"
	"log"
	"sync"

	"github.com/Zmey56/poloniex-collector/internal/domain/models"
)

type WorkerPool struct {
	numWorkers int
	taskQueue  chan *models.RecentTrade
	processor  *KlineProcessor
	wg         sync.WaitGroup
}

func NewWorkerPool(numWorkers int, processor *KlineProcessor) *WorkerPool {
	return &WorkerPool{
		numWorkers: numWorkers,
		taskQueue:  make(chan *models.RecentTrade, 1000),
		processor:  processor,
	}
}

func (wp *WorkerPool) Start(ctx context.Context) {
	for i := 0; i < wp.numWorkers; i++ {
		wp.wg.Add(1)
		go wp.worker(ctx)
	}
}

func (wp *WorkerPool) Stop() {
	close(wp.taskQueue)
	wp.wg.Wait()
}

func (wp *WorkerPool) Submit(trade *models.RecentTrade) bool {
	select {
	case wp.taskQueue <- trade:
		return true
	default:
		return false
	}
}

func (wp *WorkerPool) worker(ctx context.Context) {
	defer wp.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case trade, ok := <-wp.taskQueue:
			if !ok {
				return
			}

			if err := wp.processor.ProcessTrade(ctx, trade); err != nil {
				log.Println("error processing trade:", err)
				continue
			}
		}
	}
}
