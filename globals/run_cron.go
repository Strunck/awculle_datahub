package globals

import (
	"fmt"
	"log"

	"github.com/go-co-op/gocron/v2"
)

func (gs *GlobalState) Run_CronJob() (err error) {

	// Scheduled Tasks Rare

	// FTP Client
	if err := gs.FetchReportFiles(); err != nil {
		return fmt.Errorf("error fetching report files: %w", err)
	}
	if err := gs.DailyReportGatherer(); err != nil {
		return fmt.Errorf("error gathering daily reports: %w", err)
	}

	s, err := gocron.NewScheduler()
	defer func() { _ = s.Shutdown() }()
	if err != nil {
		return fmt.Errorf("error creating scheduler: %w", err)
	}

	s.NewJob(
		gocron.DailyJob(
			1,
			gocron.NewAtTimes(
				gocron.NewAtTime(2, 2, 2),
			),
		),
		gocron.NewTask(
			gs.TaskReportDaily(),
		),
	)

	s.Start()

	return nil
}

func (gs *GlobalState) TaskReportDaily() func() {
	return func() {
		log.Println("📃 Entering Cron Job...")

		gs.FetchReportFiles()    //ftp_client.go
		gs.DailyReportGatherer() //report.go

		log.Println("📃 Cron Job finished...")

	}
}
