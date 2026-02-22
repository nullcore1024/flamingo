package mysql

import (
	"github.com/flamingo/server/internal/base"
	"go.uber.org/zap"
	"sync"
)

type MysqlWorker struct {
	id       int
	queue    *MysqlTaskQueue
	manager  *MysqlManager
	stopChan chan struct{}
	wg       sync.WaitGroup
}

func NewMysqlWorker(id int, queue *MysqlTaskQueue, manager *MysqlManager) *MysqlWorker {
	return &MysqlWorker{
		id:       id,
		queue:    queue,
		manager:  manager,
		stopChan: make(chan struct{}),
	}
}

func (mw *MysqlWorker) Start() {
	mw.wg.Add(1)
	go mw.run()
}

func (mw *MysqlWorker) Stop() {
	close(mw.stopChan)
	mw.wg.Wait()
}

func (mw *MysqlWorker) run() {
	defer mw.wg.Done()

	base.GetLogger().Info("MySQL worker started", zap.Int("workerId", mw.id))

	for {
		select {
		case <-mw.stopChan:
			base.GetLogger().Info("MySQL worker stopped", zap.Int("workerId", mw.id))
			return
		default:
			task, ok := mw.queue.Pop()
			if !ok {
				// 队列已关闭且为空
				return
			}

			if task == nil {
				// 队列为空，继续等待
				continue
			}

			mw.processTask(task)
		}
	}
}

func (mw *MysqlWorker) processTask(task *MysqlTask) {
	var result interface{}
	var err error

	switch task.Type {
	case TaskTypeQuery:
		rows, e := mw.manager.Query(task.Query, task.Args...)
		if e != nil {
			err = e
		} else {
			defer rows.Close()
			// 这里可以根据需要处理rows，将结果转换为合适的格式
			result = rows
		}

	case TaskTypeExec:
		r, e := mw.manager.Exec(task.Query, task.Args...)
		if e != nil {
			err = e
		} else {
			result = r
		}

	case TaskTypeTransaction:
		// 处理事务
		// 这里需要根据具体的事务逻辑来实现
		base.GetLogger().Warn("Transaction task not implemented")

	default:
		base.GetLogger().Warn("Unknown task type", zap.Int("type", int(task.Type)))
	}

	// 调用回调函数
	if task.Callback != nil {
		task.Callback(result, err)
	}
}

type MysqlWorkerPool struct {
	workers []*MysqlWorker
	queue   *MysqlTaskQueue
	manager *MysqlManager
	mu      sync.Mutex
	started bool
}

func NewMysqlWorkerPool(workerCount int, queueCapacity int, manager *MysqlManager) *MysqlWorkerPool {
	queue := NewMysqlTaskQueue(queueCapacity)
	workers := make([]*MysqlWorker, workerCount)

	for i := 0; i < workerCount; i++ {
		workers[i] = NewMysqlWorker(i+1, queue, manager)
	}

	return &MysqlWorkerPool{
		workers: workers,
		queue:   queue,
		manager: manager,
		started: false,
	}
}

func (mp *MysqlWorkerPool) Start() {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	if mp.started {
		return
	}

	for _, worker := range mp.workers {
		worker.Start()
	}

	mp.started = true
	base.GetLogger().Info("MySQL worker pool started", zap.Int("workerCount", len(mp.workers)))
}

func (mp *MysqlWorkerPool) Stop() {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	if !mp.started {
		return
	}

	// 关闭队列
	mp.queue.Close()

	// 停止所有工作线程
	for _, worker := range mp.workers {
		worker.Stop()
	}

	mp.started = false
	base.GetLogger().Info("MySQL worker pool stopped")
}

func (mp *MysqlWorkerPool) SubmitTask(task *MysqlTask) error {
	return mp.queue.Push(task)
}

func (mp *MysqlWorkerPool) IsStarted() bool {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	return mp.started
}
