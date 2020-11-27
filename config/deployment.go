package config

type Environment struct {
	Name string `json:"name"`
	IAAS string `json:"iaas"`
}

type Deployment struct {
	URI          string        `json:"uri"`
	Environments []Environment `json:"environments"`
	Branch       string        `json:"branch"`
}
