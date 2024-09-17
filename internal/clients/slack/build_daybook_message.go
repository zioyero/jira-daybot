package slack

import (
	"fmt"

	"github.com/zioyero/go-slack"
	slackapi "github.com/zioyero/go-slack"
	"github.com/zioyero/jira-daybot/internal/daybook"
)

func (c *Client) buildDaybookMessage(db *daybook.Daybook) []slack.Block {
	statusUpdates := map[string]string{
		"In Progress": "Working on",
		"Code Review": "In Code Review",
		"Testing":     "Testing",
		"Done":        "Completed Today",
	}

	order := []string{"Done", "In Progress", "Code Review", "Testing"}

	bugs := daybook.TasksByStatus(db.Bugs)

	blocks := make([]slackapi.Block, 0)

	blocks = append(blocks,
		slackapi.NewSectionBlock(
			slackapi.NewTextBlockObject("mrkdwn",
				fmt.Sprintf("<@%s> *Daybook for %s*", db.User.SlackHandle, db.Day.Format("2006-01-02")), false, false,
			),
			nil,
			nil,
		),
	)

	for _, status := range order {
		epics := db.Projects[status]
		bb := bugs[status]

		taskCount := len(epics) + len(bb)

		if taskCount == 0 {
			continue
		}

		blocks = append(blocks,
			slackapi.NewSectionBlock(
				slackapi.NewTextBlockObject("mrkdwn",
					fmt.Sprintf("*%s*", statusUpdates[status]), false, false,
				),
				nil,
				nil,
			),
		)

		for _, bug := range bb {
			blocks = append(blocks, c.formatBugReport(bug, 0))
		}

		for _, epic := range epics {
			blocks = append(blocks, c.formatEpicReport(epic, 0))
		}
	}

	return blocks
}

func (c *Client) formatBugReport(bug *daybook.Task, indent int) slackapi.Block {
	return slackapi.NewRichTextBlock("",
		slackapi.NewRichTextList(slackapi.RTEListBullet, indent,
			slackapi.NewRichTextSection(
				slackapi.NewRichTextSectionEmojiElement("jira-bug", 0, nil),
				slackapi.NewRichTextSectionLinkElement(bug.Link.String(), " "+bug.Title, nil),
			),
		),
	)
}

func (c *Client) formatEpicReport(epic *daybook.Epic, indent int) *slackapi.RichTextBlock {
	listHeader := []slackapi.RichTextSectionElement{
		slackapi.NewRichTextSectionEmojiElement("jira-epic", 0, nil),
		slackapi.NewRichTextSectionLinkElement(epic.Link.String(), " "+epic.Title, nil),
	}

	storyReports := make([]slackapi.RichTextElement, 0)
	for _, story := range epic.Stories {
		storyReports = append(storyReports, c.formatStoryReport(story, indent+1)...)
	}

	contents := []slackapi.RichTextElement{
		slackapi.NewRichTextList(slackapi.RTEListBullet, indent,
			slackapi.NewRichTextSection(listHeader...),
		),
	}
	contents = append(contents, storyReports...)

	return slackapi.NewRichTextBlock("",
		contents...,
	)
}

func (c *Client) formatStoryReport(story *daybook.Story, indent int) []slackapi.RichTextElement {
	listHeader := []slackapi.RichTextSectionElement{
		slackapi.NewRichTextSectionEmojiElement("jira-story", 0, nil),
		slackapi.NewRichTextSectionLinkElement(story.Link.String(), " "+story.Title, nil),
	}

	contents := []slackapi.RichTextElement{
		slackapi.NewRichTextSection(listHeader...),
	}

	return []slackapi.RichTextElement{
		slackapi.NewRichTextList(slackapi.RTEListBullet, indent, contents...),
		c.formatSubtaskReport(story.Subtasks, indent+1),
	}
}

func (c *Client) formatSubtaskReport(subtasks []*daybook.Task, indent int) slackapi.RichTextElement {
	listItems := make([]slackapi.RichTextElement, 0)
	for _, task := range subtasks {
		listItems = append(listItems, slackapi.NewRichTextSection(
			slackapi.NewRichTextSectionEmojiElement("jira-subtask", 0, nil),
			slackapi.NewRichTextSectionLinkElement(task.Link.String(), " "+task.Title, nil),
		))
	}

	return slackapi.NewRichTextList(slackapi.RTEListBullet, indent, listItems...)
}

func (c *Client) formatPlannedTask(task *daybook.Task, _ int) slackapi.Block {
	return slackapi.NewRichTextBlock("",
		slackapi.NewRichTextSection(
			slackapi.NewRichTextSectionTextElement(task.Title, nil),
		),
	)
}
