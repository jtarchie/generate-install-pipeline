package config

import (
	"fmt"
	"github.com/concourse/concourse/atc"
	"github.com/jtarchie/generate-install-pipeline/resources"
	"strings"
)

type StepOpsManager struct {
	Version string `yaml:"version"`
}
type StepTile struct {
	Slug    string `yaml:"slug"`
	Version string `yaml:"version"`
}
type Step struct {
	OpsManager *StepOpsManager `yaml:"opsmanager"`
	Tile       *StepTile       `yaml:"tile"`
}

func (s Step) ResourceName() string {
	name := ""
	switch {
	case s.OpsManager != nil:
		name = fmt.Sprintf("opsmanager-%s", s.OpsManager.Version)
	case s.Tile != nil:
		name = fmt.Sprintf("tile-%s-%s", s.Tile.Slug, s.Tile.Version)
	}

	name = strings.ReplaceAll(name, "*", "x")
	return name
}

func (s Step) AsPivnetResource() resources.PivnetResource {
	pivnetResource := resources.PivnetResource{}
	pivnetResource.Name = s.ResourceName()

	switch {
	case s.OpsManager != nil:
		pivnetResource.Slug = "opsmanager"
		pivnetResource.Version = s.OpsManager.Version
	case s.Tile != nil:
		pivnetResource.Slug = s.Tile.Slug
		pivnetResource.Version = s.Tile.Version
	}

	return pivnetResource
}

func (s Step) AsResourceConfig() atc.ResourceConfig {
	return s.AsPivnetResource().AsResourceConfig()
}

type Steps []Step
