package mysql

import (
	"context"
	"sync"
	"time"
)

type MysqlTaskType int

const (
	TaskTypeQuery MysqlTaskType = iota
	TaskTypeExec
	TaskTypeTransaction
)

type MysqlTask struct {
	Type      MysqlTaskType
	Query     string
	Args      []interface{}
	Callback  func(result interface{}, err error)
	Ctx       context.Context
}

func NewMysqlTask(taskType MysqlTaskType, query string, args []interface{}, callback func(result interface{}, err error)) *MysqlTask {
	return &MysqlTask{
		Type:     taskType,
		Query:    query,
		Args:     args,
		Callback: callback,
		Ctx:      context.Background(),
	}
}

type MysqlTaskQueue struct {
	tasks    chan *MysqlTask
	mu       sync.Mutex
	closed   bool
}

func NewMysqlTaskQueue(capacity int) *MysqlTaskQueue {
	return &MysqlTaskQueue{
		tasks:  make(chan *MysqlTask, capacity),
		closed: false,
	}
}

func (mq *MysqlTaskQueue) Push(task *MysqlTask) error {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	if mq.closed {
		return nil
	}

	select {
	case mq.tasks <- task:
		return nil
	case <-time.After(1 * time.Second):
		return nil
	}
}

func (mq *MysqlTaskQueue) Pop() (*MysqlTask, bool) {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	if mq.closed && len(mq.tasks) == 0 {
		return nil, false
	}

	select {
	case task, ok := <-mq.tasks:
		if !ok {
			return nil, false
		}
		return task, true
	default:
		return nil, true
	}
}

func (mq *MysqlTaskQueue) Close() {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	if !mq.closed {
		mq.closed = true
		close(mq.tasks)
	}
}

func (mq *MysqlTaskQueue) Len() int {
	return len(mq.tasks)
}

func (mq *MysqlTaskQueue) IsClosed() bool {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	return mq.closed
}
