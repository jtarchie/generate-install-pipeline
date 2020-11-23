package config

import (
	"fmt"
	"strings"

	"github.com/jtarchie/generate-install-pipeline/resources"
)

type StepOpsManager struct {
	Version string `json:"version"`
}

type StepTile struct {
	Slug    string `json:"slug"`
	Version string `json:"version"`
}

type Step struct {
	OpsManager *StepOpsManager `json:"ops-manager"`
	Tile       *StepTile       `json:"tile"`
}

func (s Step) ResourceName() string {
	name := ""

	switch {
	case s.OpsManager != nil:
		name = fmt.Sprintf("ops-manager-%s", s.OpsManager.Version)
	case s.Tile != nil:
		name = fmt.Sprintf("tile-%s-%s", s.Tile.Slug, s.Tile.Version)
	}

	name = strings.ReplaceAll(name, "*", "x")

	return name
}

func (s Step) AsPivnetResource(env Environment) resources.PivnetResource {
	pivnetResource := resources.PivnetResource{}
	pivnetResource.Name = s.ResourceName()

	switch {
	case s.OpsManager != nil:
		pivnetResource.Slug = "ops-manager"
		pivnetResource.Version = s.OpsManager.Version

		switch env.IAAS {
		case "gcp", "aws", "azure":
			pivnetResource.Globs = []string{fmt.Sprintf("*%s*.yml", env.IAAS)}
		case "openstack":
			pivnetResource.Globs = []string{"*openstack*.raw"}
		case "vsphere":
			pivnetResource.Globs = []string{"*vsphere*.ova"}
		}
	case s.Tile != nil:
		pivnetResource.Slug = s.Tile.Slug
		pivnetResource.Version = s.Tile.Version
	}

	return pivnetResource
}

type Steps []Step
