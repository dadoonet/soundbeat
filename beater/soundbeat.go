package beater

import (
	"fmt"
	"time"

	"math"

	"github.com/krig/go-sox"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/publisher"

	"github.com/dadoonet/soundbeat/config"
)

type Soundbeat struct {
	done   chan struct{}
	config config.Config
	client publisher.Client
  name   string
}

// Creates beater
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	config := config.DefaultConfig
	if err := cfg.Unpack(&config); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &Soundbeat{
		done: make(chan struct{}),
		config: config,
	}

	if !sox.Init() {
		return nil, fmt.Errorf("Failed to initialize SoX")
	}
	defer sox.Quit()

  name := config.Name
  if name == "" {
		return nil, fmt.Errorf("no name set")
  }
  bt.name = name

	return bt, nil
}

func (bt *Soundbeat) Run(b *beat.Beat) error {
	logp.Info("soundbeat is running! Hit CTRL-C to stop it.")

	bt.client = b.Publisher.Connect()

	in := sox.OpenRead(bt.name)
	if in == nil {
	  logp.Err("Failed to open input file")
	}
	// Close the file before exiting
	defer in.Release()

	duration := float64(in.Signal().Length()) / float64(in.Signal().Channels()) / in.Signal().Rate()

	now := time.Now()

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

		event := common.MapStr{
      "@timestamp":   common.Time(now),
      "type":         b.Name,
      "left":         left * 100.0,
      "right":        right * 100.0,
		}
		bt.client.PublishEvent(event)
		logp.Info("Event sent")
 	}

 	logp.Info("soundbeat ended analyzing file %s", bt.name)
  return nil
}

func (bt *Soundbeat) Stop() {
	bt.client.Close()
	close(bt.done)
}
