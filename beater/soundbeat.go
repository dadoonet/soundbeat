package beater

import (
	"fmt"
	"time"

	"github.com/krig/go-sox"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/cfgfile"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"

	"github.com/dadoonet/soundbeat/config"
)

type Soundbeat struct {
	beatConfig *config.Config
	done       chan struct{}

	name   string
	period time.Duration
	zoom   float64
}

// Creates beater
func New() *Soundbeat {
	return &Soundbeat{
		done: make(chan struct{}),
	}
}

/// *** Beater interface methods ***///

func (bt *Soundbeat) Config(b *beat.Beat) error {

	// Load beater beatConfig
	err := cfgfile.Read(&bt.beatConfig, "")
	if err != nil {
		return fmt.Errorf("Error reading config file: %v", err)
	}

	return nil
}

func (bt *Soundbeat) Setup(b *beat.Beat) error {
	// All libSoX applications must start by initializing the SoX library
	if !sox.Init() {
		logp.Critical("Failed to initialize SoX")
	}
	// Make sure to call Quit before terminating
	defer sox.Quit()

	period, err := configDuration(bt.beatConfig.Soundbeat.Period, 10*time.Millisecond)
	if err != nil {
		return err
	}

	name := bt.beatConfig.Soundbeat.Name
	if name == "" {
		logp.Critical("no name set")
		return nil
	}

	zoom := 1.0
	if bt.beatConfig.Soundbeat.Zoom != 0.0 {
		zoom = bt.beatConfig.Soundbeat.Zoom
	}

	bt.name = name
	bt.period = period
	bt.zoom = zoom

	logp.Info("soundbeat has been configured:")
	logp.Info(" - Name: %s", bt.name)
	logp.Info(" - Period: %s", bt.period.String())
	logp.Info(" - Zoom: %f", bt.zoom)

	return nil
}

func configDuration(cfg string, d time.Duration) (time.Duration, error) {
	if cfg != "" {
		return time.ParseDuration(cfg)
	} else {
		return d, nil
	}
}

func (bt *Soundbeat) Run(b *beat.Beat) error {
	logp.Info("soundbeat is running! Hit CTRL-C to stop it.")

	ticker := time.NewTicker(bt.period)
	counter := 1
	for {
		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
		}

		event := common.MapStr{
			"@timestamp": common.Time(time.Now()),
			"type":       b.Name,
			"counter":    counter,
		}
		b.Events.PublishEvent(event)
		logp.Info("Event sent")
		counter++
	}
}

func (bt *Soundbeat) Cleanup(b *beat.Beat) error {
	return nil
}

func (bt *Soundbeat) Stop() {
	close(bt.done)
}
