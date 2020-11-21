package resources

import "github.com/concourse/concourse/atc"

type Resource struct {
	Name string
}

func (r Resource) AsGetStep() atc.Step {
	return atc.Step{
		Config: &atc.GetStep{
			Name: r.Name,
		},
	}
}
