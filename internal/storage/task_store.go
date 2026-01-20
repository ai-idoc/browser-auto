// Package storage 提供数据存储接口
package storage

import (
	"context"
	"sync"

	"github.com/browser-automation/internal/domain"
)

// TaskStore 任务存储接口
type TaskStore interface {
	Create(ctx context.Context, task *domain.Task) error
	Get(ctx context.Context, id string) (*domain.Task, error)
	Update(ctx context.Context, task *domain.Task) error
	UpdateStatus(ctx context.Context, id string, status domain.TaskStatus) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*domain.Task, error)
}

// MemoryTaskStore 内存任务存储（开发用）
type MemoryTaskStore struct {
	tasks map[string]*domain.Task
	mu    sync.RWMutex
}

// NewMemoryTaskStore 创建内存任务存储
func NewMemoryTaskStore() *MemoryTaskStore {
	return &MemoryTaskStore{
		tasks: make(map[string]*domain.Task),
	}
}

// Create 创建任务
func (s *MemoryTaskStore) Create(ctx context.Context, task *domain.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks[task.ID] = task
	return nil
}

// Get 获取任务
func (s *MemoryTaskStore) Get(ctx context.Context, id string) (*domain.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	task, ok := s.tasks[id]
	if !ok {
		return nil, ErrNotFound
	}
	return task, nil
}

// Update 更新任务
func (s *MemoryTaskStore) Update(ctx context.Context, task *domain.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tasks[task.ID]; !ok {
		return ErrNotFound
	}
	s.tasks[task.ID] = task
	return nil
}

// UpdateStatus 更新任务状态
func (s *MemoryTaskStore) UpdateStatus(ctx context.Context, id string, status domain.TaskStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	task, ok := s.tasks[id]
	if !ok {
		return ErrNotFound
	}
	task.Status = status
	return nil
}

// Delete 删除任务
func (s *MemoryTaskStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tasks, id)
	return nil
}

// List 列出任务
func (s *MemoryTaskStore) List(ctx context.Context, limit, offset int) ([]*domain.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	tasks := make([]*domain.Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}
	
	// 简单分页
	if offset >= len(tasks) {
		return []*domain.Task{}, nil
	}
	end := offset + limit
	if end > len(tasks) {
		end = len(tasks)
	}
	return tasks[offset:end], nil
}
