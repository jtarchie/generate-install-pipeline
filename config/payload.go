package config

type Payload struct {
	Steps      Steps      `json:"steps"`
	Deployment Deployment `json:"deployment"`
}
