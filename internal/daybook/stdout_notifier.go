package daybook

import (
	"context"
	"strings"

	"github.com/fatih/color"
)

type StdoutNotifier struct {
}

const indentAmount = 4

func (s *StdoutNotifier) SendDaybookEntry(ctx context.Context, db *Daybook) error {
	color.White("Would send daybook entry")

	color.White("@%s's Daybook for %s", db.User.SlackHandle, db.Day.Format("2006-01-02"))

	statusUpdates := map[string]string{
		"In Progress": "Working on",
		"Code Review": "In Code Review",
		"Testing":     "Testing",
		"Done":        "Completed Today",
	}

	order := []string{"Done", "In Progress", "Code Review", "Testing"}

	bugs := TasksByStatus(db.Bugs)
	standalones := TasksByStatus(db.StandaloneTasks)

	for _, status := range order {
		epics := db.Projects[status]
		bb := bugs[status]
		standalone := standalones[status]

		taskCount := len(epics) + len(bb) + len(standalone)

		if taskCount == 0 {
			continue
		}

		color.Green(statusUpdates[status])

		for _, bug := range bb {
			color.Yellow(s.formatBugReport(bug, indentAmount))
		}

		for _, epic := range epics {
			color.White(s.formatEpicReport(epic, indentAmount))
		}

		for _, task := range standalone {
			color.White(s.formatTaskReport(task, indentAmount))
		}
	}

	if len(db.CreatedTasks) > 0 {
		color.Green("Planned Tasks")
		for _, task := range db.CreatedTasks {
			color.White(s.formatPlannedTask(task, indentAmount))
		}
	}

	return nil
}

func (s *StdoutNotifier) SendDaybookDMReminder(ctx context.Context, daybook *Daybook) error {
	color.White("Would send daybook DM reminder for @%s", daybook.User.SlackHandle)

	return nil
}

func (s *StdoutNotifier) formatBugReport(bug *Task, indent int) string {
	sb := strings.Builder{}
	sb.WriteString(strings.Repeat(" ", indent))
	sb.WriteString("- :jira-bug: " + bug.Title)

	return sb.String()
}

func (s *StdoutNotifier) formatEpicReport(epic *Epic, indent int) string {
	sb := strings.Builder{}
	sb.WriteString(strings.Repeat(" ", indent))
	sb.WriteString("- :jira-epic: " + epic.Title)
	sb.WriteString("\n")

	for _, story := range epic.Stories {
		sb.WriteString(s.formatStoryReport(story, indent+indentAmount))
	}

	return sb.String()
}

func (s *StdoutNotifier) formatStoryReport(story *Story, indent int) string {
	sb := strings.Builder{}
	sb.WriteString(strings.Repeat(" ", indent))
	sb.WriteString("- :jira-story: " + story.Title)
	sb.WriteString("\n")

	for _, subtask := range story.Subtasks {
		sb.WriteString(s.formatSubtaskReport(subtask, indent+indentAmount))
		sb.WriteString("\n")
	}

	return sb.String()
}

func (s *StdoutNotifier) formatTaskReport(task *Task, indent int) string {
	sb := strings.Builder{}
	sb.WriteString(strings.Repeat(" ", indent))
	sb.WriteString("- :jira-task: " + task.Title)

	return sb.String()
}

func (s *StdoutNotifier) formatSubtaskReport(task *Task, indent int) string {
	sb := strings.Builder{}
	sb.WriteString(strings.Repeat(" ", indent))
	sb.WriteString("- :jira-subtask: " + task.Title)

	return sb.String()
}

func (s *StdoutNotifier) formatPlannedTask(task *Task, indent int) string {
	sb := strings.Builder{}
	sb.WriteString(strings.Repeat(" ", indent))
	sb.WriteString("- [Created]" + task.Title)

	return sb.String()
}
