package beater

import (
	"fmt"
	"time"
	"math"

	"github.com/krig/go-sox"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"

	"github.com/dadoonet/soundbeat/config"
)

type Soundbeat struct {
	done   chan struct{}
	config config.Config
	client beat.Client
}

// Creates beater
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	c := config.DefaultConfig
	if err := cfg.Unpack(&c); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

  if c.Name == "" {
		return nil, fmt.Errorf("no name set")
  }

	if !sox.Init() {
		return nil, fmt.Errorf("Failed to initialize SoX")
	}
	defer sox.Quit()

	bt := &Soundbeat{
		done:   make(chan struct{}),
		config: c,
	}
	return bt, nil
}

func (bt *Soundbeat) Run(b *beat.Beat) error {
	logp.Info("soundbeat ended analyzing file %s", bt.config.Name)

	var err error
	bt.client, err = b.Publisher.Connect()
	if err != nil {
		return err
	}

	in := sox.OpenRead(bt.config.Name)
	if in == nil {
	  logp.Err("Failed to open input file")
	}
	// Close the file before exiting
	defer in.Release()

	duration := float64(in.Signal().Length()) / float64(in.Signal().Channels()) / in.Signal().Rate()
	now := time.Now()

	// Rewind time so Kibana won't have to look at data in the future
	now = now.Add(-(time.Duration(duration)*time.Second))

	logp.Info("Duration of the track %s", time.Duration(duration)*time.Second)
	logp.Info("Changed starting timestamp to %s", now)

  // number of samples:
  block_size := int64(bt.config.Period.Seconds() * float64(in.Signal().Rate()) * float64(in.Signal().Channels()) + 0.5)
  // Adjust boundary:
  block_size -= block_size % int64(in.Signal().Channels())
  // Allocate memory:
  buf := make([]sox.Sample, block_size)

	for blocks := 0; in.Read(buf, uint(block_size)) == block_size && float64(blocks) * bt.config.Period.Seconds() < duration; blocks++ {
  
    left := 0.0
    right := 0.0

    now = now.Add(bt.config.Period)

    for i := int64(0); i < block_size; i++ {
      sample := sox.SampleToFloat64(buf[i])

      if (i & 1) != 0 {
        right = math.Max(right, math.Abs(sample))
      } else {
        left = math.Max(left, math.Abs(sample))
      }
    }

		event := beat.Event{
			Timestamp: now,
			Fields: common.MapStr{
				"type":       bt.config.Name,
				"left":       left * 100.0,
				"right":      right * 100.0,
			},
		}
		bt.client.Publish(event)
		logp.Debug("Event sent", "")
	}

	logp.Info("soundbeat ended analyzing file %s", bt.config.Name)
	return nil
}

func (bt *Soundbeat) Stop() {
	bt.client.Close()
	close(bt.done)
}
