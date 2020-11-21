package resources

import (
	"github.com/concourse/concourse/atc"
	"strings"
)

type PivnetResource struct {
	Resource

	Slug    string
	Version string
	Globs   []string
}

func (p PivnetResource) AsResourceConfig() atc.ResourceConfig {
	source := map[string]interface{}{
		"api_token":       "((pivnet.api_token))",
		"product_slug":    p.Slug,
		"product_version": strings.ReplaceAll(p.Version, ".", "\\."),
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
	if len(p.Globs) > 0 {
		step.Params["globs"] = p.Globs
	}
	return atc.Step{
		Config: &step,
	}
}
