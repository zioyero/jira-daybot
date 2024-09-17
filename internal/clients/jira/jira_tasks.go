package jira

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/zioyero/jira-daybot/internal/daybook"
)

// UserTasks returns all tasks assigned to the user that are in progress or in code review, as well
// as tasks that have marked as done in the last 24 hours, in order to populate the daybook.
func (c *Client) UserTasks(ctx context.Context, user *daybook.User) ([]*daybook.Task, error) {
	slog.Info("Getting user tasks")

	query := fmt.Sprintf("project = %s AND type != EPIC AND (assignee IN (%q)) AND ((status IN (\"In Progress\", \"Code Review\", \"Testing\")) OR (status IN (\"Done\") AND updated >= -24h) OR (status = \"To Do\" AND updated >= -24h))", c.cfg.Project, user.AtlassianID)

	issues, _, err := c.jira.Issue.Search(ctx, query, nil)
	if err != nil {
		return nil, fmt.Errorf("searching issues: %w", err)
	}

	return c.unmarshalTasks(issues)
}

func (c *Client) Task(ctx context.Context, taskID string) (*daybook.Task, error) {
	issue, _, err := c.jira.Issue.Get(ctx, taskID, nil)
	if err != nil {
		return nil, fmt.Errorf("getting issue: %w", err)
	}

	if issue == nil {
		return nil, fmt.Errorf("issue not found")
	}

	return c.unmarshalTask(*issue)
}

// RootTask returns the root task for the given task ID. If the task has a parent, it will recursively
// call itself until it finds the root task.
func (c *Client) RootTask(ctx context.Context, taskID string) (*daybook.Task, error) {
	slog.Info(fmt.Sprintf("Getting root task for %s", taskID))

	issue, _, err := c.jira.Issue.Get(ctx, taskID, nil)
	if err != nil {
		return nil, fmt.Errorf("getting issue: %w", err)
	}

	if issue.Fields.Parent == nil {
		return c.unmarshalTask(*issue)
	}

	return c.RootTask(ctx, issue.Fields.Parent.Key)
}

// CreatedByUser returns all tasks created by the user since the beginning of the day,
// according to the JQL startOfDay() function.
func (c *Client) CreatedByUser(ctx context.Context) ([]*daybook.Task, error) {
	slog.Info("Getting tasks created by user")

	query := fmt.Sprintf("project = %s AND reporter = currentUser() and created >= startOfDay()", c.cfg.Project)

	issues, _, err := c.jira.Issue.Search(ctx, query, nil)
	if err != nil {
		return nil, fmt.Errorf("searching issues: %w", err)
	}

	return c.unmarshalTasks(issues)
}
