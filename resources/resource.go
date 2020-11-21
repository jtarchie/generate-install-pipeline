package resources

import "github.com/concourse/concourse/atc"

type Resource struct {
	Name  string
	Alias string
}

func (r Resource) AsGetStep() atc.Step {
	step := &atc.GetStep{
		Name: r.Name,
	}

	if r.Alias != "" {
		step.Resource = r.Alias
	}

	return atc.Step{
		Config: step,
	}
}
