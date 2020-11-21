package config

type Deployment struct {
	URI          string `yaml:"uri"`
	Environments []struct {
		Name string `yaml:"name"`
		IAAS string `yaml:"iaas"`
	} `yaml:"environments"`
}
