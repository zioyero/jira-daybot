package daybook

import (
	"net/url"
	"time"
)

type Daybook struct {
	Day             time.Time
	User            *User
	Projects        map[string][]*Epic
	Bugs            []*Task
	CreatedTasks    []*Task
	StandaloneTasks []*Task
}

type User struct {
	SlackHandle     string
	SlackID         string
	AtlassianID     string
	DaybookChannels []string
}

type Task struct {
	Type         string
	ID           string
	Status       string
	Link         *url.URL
	Title        string
	ParentTaskID string
}

type Epic struct {
	*Task

	Stories []*Story
}

// ContainsTask checks if the epic contains a task with the given ID.
// A task is considered to be contained in an epic if it is either a story or a subtask of a story
// in the epic.
func (e *Epic) ContainsTask(taskID string) bool {
	for _, story := range e.Stories {
		if story.ID == taskID {
			return true
		}

		for _, subtask := range story.Subtasks {
			if subtask.ID == taskID {
				return true
			}
		}
	}

	return false
}

type Story struct {
	*Task

	Subtasks []*Task
	Epic     *Epic
}

type PullRequest struct {
	Title string
	Link  *url.URL
}
