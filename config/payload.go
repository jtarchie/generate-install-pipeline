package config

type Payload struct {
	Steps       Steps       `yaml:"steps"`
	Deployments Deployments `yaml:"deployments"`
}
