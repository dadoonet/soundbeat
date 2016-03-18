// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

type Config struct {
	Soundbeat SoundbeatConfig
}

type SoundbeatConfig struct {
	Name   string  `yaml:"name"`
	Period string  `yaml:"period"`
	Zoom   float64 `yaml:"zoom"`
}
