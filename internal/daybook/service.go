package daybook

import (
	"context"
)

type Notifier interface {
	SendDaybookEntry(ctx context.Context, daybook *Daybook) error
	SendDaybookDMReminder(ctx context.Context, daybook *Daybook) error
}

type TaskRepository interface {
	UserTasks(ctx context.Context, user *User) ([]*Task, error)
	Task(ctx context.Context, taskID string) (*Task, error)
	RootTask(ctx context.Context, taskID string) (*Task, error)
	CreatedByUser(ctx context.Context) ([]*Task, error)
}

type Config struct {
}

type Service struct {
	cfg      Config
	notifier Notifier
	tasks    TaskRepository
}

func NewService(cfg Config, notifier Notifier, tasks TaskRepository) *Service {
	return &Service{
		cfg:      cfg,
		notifier: notifier,
		tasks:    tasks,
	}
}
