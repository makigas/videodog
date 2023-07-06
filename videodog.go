package main

import (
	"errors"
	"flag"
	"log"
	"os"
	"time"

	"github.com/go-co-op/gocron"
)

var (
	config *Config // The execution context of the application.
	spool  *Spool  // The spool used to store the system data.

	flagConfig     string // The config file that should be used.
	flagNoannounce bool   // Whether to avoid announcing the videos (part of dry-run mode)
)

// The different errors used by the application.
var (
	ErrNoConfigFile        = errors.New("missing config file")
	ErrInvalidConfigFile   = errors.New("invalid config file")
	ErrSpoolInitialization = errors.New("error initializing spool")
	ErrSpoolOperation      = errors.New("error operating spool")
	ErrSchedulerConfig     = errors.New("cannot configure scheduler")
	ErrFeedFetcher         = errors.New("feed fetching error")
	ErrAnnouncer           = errors.New("announce hook error")
)

func main() {
	// Startup application.
	parseArgs()
	initContext()
	defer spool.Close()

	// If the configuration says that we must autodiscard entries on start, do it now.
	if config.Autospool {
		autospoolFeeds()
	}

	// Exec daemon until we receive SIGINT.
	startDaemon()
}

func initContext() {
	file, err := os.Open(flagConfig)
	if err != nil {
		gentlePanic(errors.Join(ErrInvalidConfigFile, err))
	}
	defer file.Close()

	config, err = ReadConfigObject(file)
	if err != nil {
		gentlePanic(errors.Join(ErrInvalidConfigFile, err))
	}

	spool, err = NewSpool(config.Spoolfile)
	if err != nil {
		gentlePanic(errors.Join(ErrSpoolInitialization, err))
	}
}

func parseArgs() {
	flag.StringVar(&flagConfig, "config", "", "The config file to use")
	flag.BoolVar(&flagNoannounce, "noannounce", false, "Don't announce the videos")
	flag.Parse()

	if flagConfig == "" {
		flag.PrintDefaults()
		os.Exit(0)
	}
}

// Fetches each feed and marks all the content in the feeds as announced.
// Prevents a ping-storm on first boot of the application. Only videos uploaded
// since the daemon starts are announced.
func autospoolFeeds() {
	for key, channel := range config.Channels {
		videos := fetch(channel)
		for _, video := range videos {
			if !announced(key, &video) {
				mark(key, &video)
			}
		}
	}
}

// Starts the scheduler and lets it block the application.
func startDaemon() {
	scheduler := gocron.NewScheduler(time.UTC)
	_, err := scheduler.Every(1).Minutes().Do(func() {
		for key, channel := range config.Channels {
			checkChannel(key, channel)
		}
	})
	if err != nil {
		gentlePanic(errors.Join(ErrSchedulerConfig, err))
	}
	scheduler.StartBlocking()
}

func checkChannel(key string, channel *Channel) {
	videos := fetch(channel)
	for _, video := range videos {
		if !announced(key, &video) {
			if flagNoannounce {
				log.Printf("Skipping announce because running in dry-mode")
			} else {
				announce(channel, &video)
			}
			mark(key, &video)
			// only one video per iteration, prevents ping-storms
			return
		}
	}
}

func gentlePanic(err error) {
	log.Panic(err)
}

func fetch(channel *Channel) []Video {
	if channel.TryCheckStale() {
		videos, err := FetchYouTubeFeed(channel.Feed)
		if err != nil {
			gentlePanic(errors.Join(ErrFeedFetcher, err))
		}
		return videos
	}
	return []Video{}
}

func announced(channel string, video *Video) bool {
	done, err := spool.IsAnnounced(channel, video.VideoId)
	if err != nil {
		gentlePanic(errors.Join(ErrSpoolOperation, err))
	}
	return done
}

func announce(channel *Channel, video *Video) {
	if err := AnnounceVideo(channel.Webhook, channel.Role, video); err != nil {
		gentlePanic(errors.Join(ErrAnnouncer, err))
	}
}

func mark(channel string, video *Video) {
	if err := spool.MarkAsAnnounced(channel, video.VideoId); err != nil {
		gentlePanic(errors.Join(ErrSpoolOperation, err))
	}
}
