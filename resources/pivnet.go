package resources

import (
	"strings"

	"github.com/concourse/concourse/atc"
)

type PivnetResource struct {
	Resource

	Slug    string
	Version string
	Globs   []string
	Unpack  bool
}

func (p PivnetResource) AsResourceConfig() atc.ResourceConfig {
	productVersion := strings.ReplaceAll(p.Version, ".", "\\.")
	productVersion = strings.ReplaceAll(productVersion, "*", ".*")

	source := map[string]interface{}{
		"api_token":       "((pivnet.api_token))",
		"product_slug":    p.Slug,
		"product_version": productVersion,
	}

	return atc.ResourceConfig{
		Name:   p.Name,
		Type:   "pivnet",
		Source: source,
	}
}

func (p PivnetResource) AsGetStep() atc.Step {
	step := atc.GetStep{
		Name:   p.Resource.Name,
		Params: map[string]interface{}{},
	}

	if p.Alias != "" {
		step.Resource = p.Alias
	}

	if len(p.Globs) > 0 {
		step.Params["globs"] = p.Globs
	}

	if p.Unpack {
		step.Params["unpack"] = true
	}

	return atc.Step{
		Config: &step,
	}
}
