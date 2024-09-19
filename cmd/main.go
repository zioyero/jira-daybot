package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/go-co-op/gocron/v2"
	"github.com/joho/godotenv"
	"github.com/zioyero/jira-daybot/internal/clients/jira"
	"github.com/zioyero/jira-daybot/internal/clients/slack"
	"github.com/zioyero/jira-daybot/internal/daybook"
)

var (
	runNowFlag = flag.Bool("run-now", false, "Run the jobs immediately upon starting")
	outputFlag = flag.String("output", "stdout", "Output destination")
)

const (
	teamPublishingEngChannel = "C04DTBQRVUK"
	engBackendChannel        = "C032E009U2C"
	devNullChannel           = "C07KPQHT7L7"
)

var users = []*daybook.User{
	{AtlassianID: "61843ea1892c420072fdd376", SlackHandle: "acastillejos", SlackID: "U02L4NL51B6", DaybookChannels: []string{devNullChannel}},
	// {AtlassianID: "630510117cfac1bfa6f9e0fb", SlackHandle: "jacob", SlackID: "U03SQC6F7L7", DaybookChannels: []string{teamPublishingEngChannel}},
}

func main() {
	flag.Parse()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	color.White("Starting JIRA Daybook Deamon")
	if *runNowFlag {
		color.Yellow("Running jobs immediately")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	d := build()
	s, jobs := scheduleJobs(ctx, d)

	s.Start()

	color.White("JIRA Daybook Daemon started")

	color.White("Configured users: %s", users)

	for _, j := range jobs {
		nextRun, err := j.NextRun()
		if err != nil {
			color.Red("Error getting next run for job %s: %v", j.Name(), err)
			os.Exit(1)
		}
		color.White("Scheduled job %s will run next at %v, %s from now", j.Name(), nextRun, nextRun.Sub(time.Now()))

		if *runNowFlag {
			err = j.RunNow()
			if err != nil {
				color.Red("Error running job %s: %v", j.Name(), err)
				os.Exit(1)
			}
		}
	}

	<-ctx.Done()

	color.Yellow("Shutting down JIRA Daybook Daemon")

	s.Shutdown()
}

func build() *daybook.Service {
	jiraTasks, err := jira.NewClient(jira.Config{
		JiraInstance: os.Getenv("JIRA_INSTANCE"),
		APIToken:     os.Getenv("JIRA_TOKEN"),
		Username:     os.Getenv("JIRA_USER"),
		Project:      os.Getenv("JIRA_PROJECT"),
	})
	if err != nil {
		color.Red("Error creating JIRA client: %v", err)
		os.Exit(1)
	}

	// Output options
	slackClient := slack.NewClient(&slack.Config{
		Token:          os.Getenv("SLACK_TOKEN"),
		DaybookChannel: os.Getenv("DAYBOOK_CHANNEL"),
	})
	stdoutNotifier := &daybook.StdoutNotifier{}

	var output daybook.Notifier
	switch *outputFlag {
	case "slack":
		output = slackClient
	case "stdout":
		output = stdoutNotifier
	default:
		color.Red("Invalid output flag: %s", *outputFlag)
		os.Exit(1)
	}

	cfg := daybook.Config{}

	daybook := daybook.NewService(cfg, output, jiraTasks)

	return daybook
}

func scheduleJobs(ctx context.Context, d *daybook.Service) (gocron.Scheduler, []gocron.Job) {
	location, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		log.Fatalf("Error loading location: %v", err)
	}

	s, err := gocron.NewScheduler(gocron.WithLocation(location))
	if err != nil {
		log.Fatalf("Error creating scheduler: %v", err)
	}

	// Send daybook entry every weekday at 4:30 PM
	entryJob, err := s.NewJob(gocron.CronJob(os.Getenv("DAYBOOK_CRONTAB"), false), gocron.NewTask(func() {
		err = d.SendDaybookEntries(ctx, users)
		if err != nil {
			color.Red("Error sending daybook entry: %v", err)
			os.Exit(1)
		}
	}), gocron.WithName("SendDaybookEntry"))
	if err != nil {
		log.Fatalf("Error creating job: %v", err)
	}

	// Send daybook reminder DM every weekday at 4:00 PM
	dmJob, err := s.NewJob(gocron.CronJob(os.Getenv("REMINDER_CRONTAB"), false), gocron.NewTask(func() {
		err = d.SendDaybookDMReminders(ctx, users)
		if err != nil {
			color.Red("Error sending daybook reminder: %v", err)
			os.Exit(1)
		}
	}), gocron.WithName("SendDaybookDMReminder"))
	if err != nil {
		log.Fatalf("Error creating job: %v", err)
	}

	return s, []gocron.Job{entryJob, dmJob}
}
