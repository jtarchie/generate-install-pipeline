package pipeline

import (
	"github.com/concourse/concourse/atc"
	"github.com/jtarchie/generate-install-pipeline/config"
	"github.com/jtarchie/generate-install-pipeline/resources"
)

type Creator struct {
	config  atc.Config
	payload config.Payload
}

func (p *Creator) setupResourceTypes() {
	p.config.ResourceTypes = append(
		p.config.ResourceTypes,
		atc.ResourceType{
			Name: "pivnet",
			Type: "registry-image",
			Source: map[string]interface{}{
				"repository": "pivotalcf/pivnet-resource",
				"tag":        "latest-final",
			},
		},
	)
}

func (p *Creator) setupDefaultResources() {
	p.config.Resources = append(
		p.config.Resources,
		p.deploymentResource().AsResourceConfig(),
		p.pavingResource().AsResourceConfig(),
		p.platformAutomationResource().AsResourceConfig(),
	)
}

func (p Creator) platformAutomationResource() resources.PivnetResource {
	return resources.PivnetResource{
		Resource: resources.Resource{Name: "platform-automation"},
		Slug:     "platform-automation",
		Version:  ".*",
	}
}

func (p Creator) platformAutomationImageResource() resources.PivnetResource {
	return resources.PivnetResource{
		Resource: resources.Resource{
			Name:  "platform-automation-image",
			Alias: "platform-automation",
		},
		Globs:  []string{"*image*.tgz"},
		Unpack: true,
	}
}

func (p Creator) platformAutomationTasksResource() resources.PivnetResource {
	return resources.PivnetResource{
		Resource: resources.Resource{
			Name:  "platform-automation-tasks",
			Alias: "platform-automation",
		},
		Globs:  []string{"*tasks*.zip"},
		Unpack: true,
	}
}

func (p Creator) deploymentResource() resources.GitResource {
	return resources.GitResource{
		Resource: resources.Resource{Name: "deployments"},
		URI:      p.payload.Deployment.URI,
	}
}

func (p Creator) pavingResource() resources.GitResource {
	return resources.GitResource{
		Resource: resources.Resource{Name: "paving"},
		URI:      "https://github.com/pivotal/paving",
	}
}

func (p *Creator) setupJobDefaultGets() {
	p.config.Jobs = append(
		p.config.Jobs,
		atc.JobConfig{
			Name:   "build",
			Serial: true,
		},
	)

	p.config.Jobs[0].PlanSequence = append(p.config.Jobs[0].PlanSequence,
		p.deploymentResource().AsGetStep(),
		p.pavingResource().AsGetStep(),
		p.platformAutomationImageResource().AsGetStep(),
		p.platformAutomationTasksResource().AsGetStep(),
	)
}

func (p *Creator) setupSteps() {
	for _, step := range p.payload.Steps {
		p.config.Jobs[0].PlanSequence = append(
			p.config.Jobs[0].PlanSequence,
			step.AsPivnetResource().AsGetStep(),
		)

		p.config.Resources = append(
			p.config.Resources,
			step.AsResourceConfig(),
		)
	}
}

func (p *Creator) AsPipeline() atc.Config {
	return p.config
}

func (p *Creator) setupTasks() {
	//p.config.Jobs[0].PlanSequence = append(
	//	p.config.Jobs[0].PlanSequence,
	//)
}

func NewCreator(c config.Payload) *Creator {
	p := Creator{
		payload: c,
	}
	p.setupResourceTypes()
	p.setupDefaultResources()
	p.setupJobDefaultGets()
	p.setupSteps()
	p.setupTasks()

	return &p
}
