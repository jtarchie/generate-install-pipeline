package resources

import (
	"fmt"
	"strings"

	"github.com/concourse/concourse/atc"
)

type GitResource struct {
	Resource

	URI    string
	Branch string
}

func (g GitResource) AsResourceConfig() atc.ResourceConfig {
	source := map[string]interface{}{}
	source["uri"] = g.URI

	if g.Branch != "" {
		source["branch"] = g.Branch
	}

	if strings.HasPrefix(g.URI, "git@") {
		source["private_key"] = fmt.Sprintf("%s.private_key", g.Name)
	}

	return atc.ResourceConfig{
		Name:   g.Name,
		Type:   "git",
		Source: source,
	}
}
