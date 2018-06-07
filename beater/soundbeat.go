package beater

import (
	"fmt"
	"math"
	"time"

	"github.com/krig/go-sox"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"

	"github.com/dadoonet/soundbeat/config"
)

// Soundbeat configuration
type Soundbeat struct {
	done   chan struct{}
	config config.Config
	client beat.Client
}

// New Creates beater
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

// Run runs the beat
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

	// This duration is in milliseconds
	duration := float64(in.Signal().Length()) / float64(in.Signal().Channels()) / in.Signal().Rate() * 1000
	start := time.Now()
	now := time.Now()
	firstTick := now
	lastTick := now
	sampleNum := 0

	logp.Info("Duration of the track %s", time.Duration(duration)*time.Millisecond)
	logp.Info("Size of the track %s", in.Signal().Length())
	// logp.Info("Changed starting timestamp to %s", now)
	logp.Info("Period is %s", bt.config.Period)
	// number of samples to read per period (including stereo or mono)
	blockSize := int64(bt.config.Period.Seconds()*float64(in.Signal().Rate())*float64(in.Signal().Channels()) + 0.5)
	logp.Info("number of samples: %s", blockSize)
	// Adjust boundary:
	blockSize -= blockSize % int64(in.Signal().Channels())
	logp.Info("number of samples after adjustment: %s", blockSize)

	// Allocate memory for each sample
	buf := make([]sox.Sample, blockSize)

	left := 0.0
	right := 0.0

	logp.Info("Start music at %s with left = %s and right = %s", now, left, right)

	for blocks := 0; in.Read(buf, uint(blockSize)) == blockSize && float64(blocks)*bt.config.Period.Seconds() < duration; blocks++ {

		left = 0
		right = 0

		for i := int64(0); i < blockSize; i++ {
			sample := sox.SampleToFloat64(buf[i])

			if (i & 1) != 0 {
				right = math.Max(right, math.Abs(sample))
			} else {
				left = math.Max(left, math.Abs(sample))
			}
		}

		sampleNum++
		event := beat.Event{
			Timestamp: now,
			Fields: common.MapStr{
				"type":   bt.config.Name,
				"sample": sampleNum,
				"left":   left * 100.0,
				"right":  right * 100.0,
			},
		}
		bt.client.Publish(event)
		logp.Debug("Sample %s sent", string(sampleNum))
		now = now.Add(bt.config.Period)
		lastTick = now
	}

	logp.Info("Stop music at %s with left = %s and right = %s", now, left, right)
	logp.Info("Music duration that we analyzed = %s", lastTick.Sub(firstTick))
	logp.Info("Number of samples generated = %s", sampleNum)

	end := time.Now()

	logp.Info("Took %s to analyze the music", end.Sub(start))

	logp.Info("soundbeat ended analyzing file %s", bt.config.Name)
	return nil
}

// Stop is called when we stop the beat
func (bt *Soundbeat) Stop() {
	bt.client.Close()
	close(bt.done)
}
