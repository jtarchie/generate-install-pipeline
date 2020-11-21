package resources

import (
	"github.com/concourse/concourse/atc"
	"strings"
)

type PivnetResource struct {
	Resource

	Slug    string
	Version string
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
