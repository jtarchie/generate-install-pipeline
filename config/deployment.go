package config

type Deployment struct {
	Name string `yaml:"name"`
	IAAS string `yaml:"iaas"`
}
type Deployments []Deployment
