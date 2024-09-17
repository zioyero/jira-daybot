package jira

import (
	"fmt"
	"net/url"

	jiralib "github.com/andygrunwald/go-jira/v2/cloud"
	"github.com/zioyero/jira-daybot/internal/daybook"
)

func (c *Client) unmarshalTasks(issues []jiralib.Issue) ([]*daybook.Task, error) {
	tasks := make([]*daybook.Task, 0, len(issues))

	for _, issue := range issues {
		t, err := c.unmarshalTask(issue)
		if err != nil {
			return nil, fmt.Errorf("unmarshaling task: %w", err)
		}
		tasks = append(tasks, t)
	}

	return tasks, nil
}

func (c *Client) unmarshalTask(issue jiralib.Issue) (*daybook.Task, error) {
	link, err := url.Parse(c.cfg.JiraInstance + "/browse/" + issue.Key)
	if err != nil {
		return nil, fmt.Errorf("parsing issue link: %w", err)
	}

	parentKey := ""
	if issue.Fields.Parent != nil {
		// Parent key is in the format "PROJECT-123", so we need to extract the number
		parentKey = issue.Fields.Parent.Key
	}

	return &daybook.Task{
		ID:           issue.Key,
		Title:        issue.Fields.Summary,
		Link:         link,
		Status:       issue.Fields.Status.Name,
		ParentTaskID: parentKey,
		Type:         issue.Fields.Type.Name,
	}, nil
}
