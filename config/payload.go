package config

type Payload struct {
	Steps      Steps      `yaml:"steps"`
	Deployment Deployment `yaml:"deployment"`
}
