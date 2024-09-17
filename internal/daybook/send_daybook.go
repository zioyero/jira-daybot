package daybook

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/fatih/color"
)

func (s *Service) SendDaybookEntries(ctx context.Context, users []*User) error {
	for _, userID := range users {
		err := s.SendDaybookEntry(ctx, userID)
		if err != nil {
			slog.Error("Sending daybook entry", "UserID", userID, "Error", err)
		}
	}

	return nil
}

// SendDayBookEntry computes the daybook entry for the current day and sends it to the notifier
func (s *Service) SendDaybookEntry(ctx context.Context, user *User) error {
	color.White("Sending daybook entry for @%s", user.SlackHandle)

	// Generate the daybook entry
	daybook, err := s.generateDaybookEntry(ctx, user)
	if err != nil {
		return fmt.Errorf("generating daybook entry: %w", err)
	}

	// Send the daybook entry to the notifier
	err = s.notifier.SendDaybookEntry(ctx, daybook)
	if err != nil {
		return fmt.Errorf("sending daybook entry: %w", err)
	}

	return nil
}

func (s *Service) SendDaybookDMReminders(ctx context.Context, users []*User) error {
	for _, user := range users {
		err := s.SendDaybookDMReminder(ctx, user)
		if err != nil {
			slog.Error("Sending daybook DM reminder", "UserID", user.SlackHandle, "Error", err)
		}
	}

	return nil
}

func (s *Service) SendDaybookDMReminder(ctx context.Context, user *User) error {
	color.White("Sending daybook DM reminder for @%s", user.SlackHandle)

	// Generate the daybook entry
	daybook, err := s.generateDaybookEntry(ctx, user)
	if err != nil {
		return fmt.Errorf("generating daybook entry: %w", err)
	}

	// Send the daybook entry to the notifier
	err = s.notifier.SendDaybookDMReminder(ctx, daybook)
	if err != nil {
		return fmt.Errorf("sending daybook DM reminder: %w", err)
	}

	return nil
}

func (s *Service) generateDaybookEntry(ctx context.Context, user *User) (*Daybook, error) {
	daybook := &Daybook{Day: time.Now(), User: user}

	// Get all the tasks assigned to the user
	tasks, err := s.tasks.UserTasks(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("getting user tasks: %w", err)
	}

	color.Green("User has %d assigned tasks", len(tasks))

	// Organize the tasks into the daybook entry

	err = s.populateProjects(ctx, daybook, tasks)
	if err != nil {
		return nil, fmt.Errorf("populating projects: %w", err)
	}

	err = s.populateBugs(ctx, daybook, tasks)
	if err != nil {
		return nil, fmt.Errorf("populating bugs: %w", err)
	}

	err = s.populateStandaloneTasks(ctx, daybook, tasks)
	if err != nil {
		return nil, fmt.Errorf("populating standalone tasks: %w", err)
	}

	return daybook, nil
}

// populateProjects takes a set of tasks and populates the Projects field of the daybook with the tasks
// organized by project. This is done by creating a tree of Epics, Stories, and Subtasks.
func (s *Service) populateProjects(ctx context.Context, daybook *Daybook, tasks []*Task) error {
	// Group the tasks by status
	byStatus := make(map[string][]*Task)
	for _, task := range tasks {
		byStatus[task.Status] = append(byStatus[task.Status], task)
	}

	// Get the task tree for each status, which places it in the correct Epic for
	// per-project reporting
	projectTasks := make(map[string][]*Epic)
	for status, tasks := range byStatus {
		tree, err := s.organizeEpics(ctx, tasks)
		if err != nil {
			return fmt.Errorf("getting task tree: %w", err)
		}

		projectTasks[status] = tree
	}

	daybook.Projects = projectTasks

	return nil
}

// organizeEpics takes a set of tasks and organizes them into a tree of Epics, Stories, and Subtasks.
// Tasks that are not within an Epic are ignored and not included in the output.
func (s *Service) organizeEpics(ctx context.Context, tasks []*Task) ([]*Epic, error) {
	// First, get the set of stories from the tasks
	stories := make(map[string]*Story)
	for _, task := range tasks {
		if task.Type == "Story" {
			stories[task.ID] = &Story{Task: task}
		}
	}

	// Next, put all the subtasks into their parent stories
	for _, task := range tasks {
		if task.Type == "Sub-task" {
			slog.Info("Getting parent task for subtask", "Subtask", task.ID)
			parent, err := s.tasks.Task(ctx, task.ParentTaskID)
			if err != nil {
				return nil, fmt.Errorf("getting parent task: %w", err)
			}

			if _, ok := stories[parent.ID]; !ok {
				stories[parent.ID] = &Story{Task: parent, Subtasks: []*Task{}}
			}

			stories[parent.ID].Subtasks = append(stories[parent.ID].Subtasks, task)
		}
	}

	// Get the set of epics from the stories
	epics := make(map[string]*Epic)
	for _, story := range stories {
		epic, err := s.tasks.RootTask(ctx, story.ID)
		if err != nil {
			return nil, fmt.Errorf("getting root task: %w", err)
		}

		if _, ok := epics[epic.ID]; !ok {
			epics[epic.ID] = &Epic{Task: epic, Stories: []*Story{}}
		}

		epics[epic.ID].Stories = append(epics[epic.ID].Stories, story)
	}

	// Convert the map to a slice
	ee := make([]*Epic, 0, len(epics))
	for _, epic := range epics {
		ee = append(ee, epic)
	}

	return ee, nil
}

// populateBugs takes a set of tasks and populates the Bugs field of the daybook with the tasks
// that are of type Bug. Bugs are not typically associated with a project, so they are not included
// in the Projects field.
func (s *Service) populateBugs(_ context.Context, daybook *Daybook, tasks []*Task) error {
	bugs := make([]*Task, 0)
	for _, task := range tasks {
		if task.Type == "Bug" {
			bugs = append(bugs, task)
		}
	}

	daybook.Bugs = bugs

	return nil
}

func (s *Service) populateStandaloneTasks(_ context.Context, daybook *Daybook, tasks []*Task) error {
	// Get the tasks without an associated project, which capture non-project work
	standaloneTasks := make([]*Task, 0)
	for _, task := range tasks {
		inProject := false
		for _, tree := range daybook.Projects {
			for _, epic := range tree {
				if epic.ContainsTask(task.ID) {
					inProject = true
					break
				}
			}
		}

		if !inProject {
			standaloneTasks = append(standaloneTasks, task)
		}
	}

	daybook.StandaloneTasks = standaloneTasks

	return nil
}

func (s *Service) populatePlannedTasks(ctx context.Context, daybook *Daybook) error {
	// Get the tasks created by the user today, since it indicates planning work
	createdTasks, err := s.tasks.CreatedByUser(ctx)
	if err != nil {
		return fmt.Errorf("getting tasks created by user: %w", err)
	}

	daybook.CreatedTasks = createdTasks

	return nil
}

func (s *Service) printEpics(title string, epics []*Epic) {
	color.White(title)
	for _, epic := range epics {
		color.White("  Epic: %s", epic.Title)
		for _, story := range epic.Stories {
			color.White("    Story: %s", story.Title)
			for _, subtask := range story.Subtasks {
				color.White("      Subtask: %s", subtask.Title)
			}
		}
	}
}
