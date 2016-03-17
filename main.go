package main

import (
	"os"

	"github.com/elastic/beats/libbeat/beat"

	"github.com/dadoonet/soundbeat/beater"
)

func main() {
	err := beat.Run("soundbeat", "", beater.New())
	if err != nil {
		os.Exit(1)
	}
}
