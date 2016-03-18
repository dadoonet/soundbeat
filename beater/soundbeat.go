package beater

import (
	"fmt"
	"time"
	"math"

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
	logp.Info("soundbeat is starting...")

	// Open the input file (with default parameters)
	in := sox.OpenRead(bt.name)
	if in == nil {
		logp.Err("Failed to open input file")
	}
	// Close the file before exiting
	defer in.Release()

	// This example program requires that the audio has precisely 2 channels:
	if in.Signal().Channels() != 2 {
		logp.Err("Input must be 2 channels")
	}

	// Total duration is: number of samples (Length()) / 2 (stereo file) / rate (Rate())
	// Example: The Whispers:
	// - Length: 26 123 340 samples
	// - Stereo: yes
	// - Rate: 44 100 Hz
	// Duration is 26123340/2/44100=296.182s

	duration := float64(in.Signal().Length()) / float64(in.Signal().Channels()) / in.Signal().Rate()
	logp.Info(" - Duration %f s", duration)
	logp.Info(" - Channels %d", in.Signal().Channels())
	logp.Info(" - Rate %d Hz", int64(in.Signal().Rate()))

	now := time.Now()

	period := bt.period

	// Convert block size (in seconds) to a number of samples:
	block_size := int64(period.Seconds()*float64(in.Signal().Rate())*float64(in.Signal().Channels()) + 0.5)
	// Make sure that this is at a `wide sample' boundary:
	block_size -= block_size % int64(in.Signal().Channels())
	// Allocate a block of memory to store the block of audio samples:
	buf := make([]sox.Sample, block_size)

	// Read and process blocks of audio for the selected duration or until EOF:
	for blocks := 0; in.Read(buf, uint(block_size)) == block_size && float64(blocks)*period.Seconds() < duration; blocks++ {
		left := 0.0
		right := 0.0

		// We increment time with a block_size (in seconds)
		// But we first zoom accordingly to the zoom factor we set
		modifed_period := float64(period.Nanoseconds()) * bt.zoom
		now = now.Add(time.Duration(modifed_period))

		for i := int64(0); i < block_size; i++ {
			// convert the sample from SoX's internal format to a `float64' for
			// processing in this application:
			sample := sox.SampleToFloat64(buf[i])

			// The samples for each channel are interleaved; in this example
			// we allow only stereo audio, so the left channel audio can be found in
			// even-numbered samples, and the right channel audio in odd-numbered
			// samples:
			if (i & 1) != 0 {
				right = math.Max(right, math.Abs(sample))
			} else {
				left = math.Max(left, math.Abs(sample))
			}
		}

		event := common.MapStr{
			"@timestamp": common.Time(now),
			"type":       b.Name,
			"left":       left * 100.0,
			"right":      right * 100.0,
		}

		b.Events.PublishEvent(event)
	}

	logp.Info("soundbeat ended analyzing file %s", bt.name)
	return nil
}

func (bt *Soundbeat) Cleanup(b *beat.Beat) error {
	return nil
}

func (bt *Soundbeat) Stop() {
	close(bt.done)
}
