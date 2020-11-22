package pipeline

import (
	"fmt"

	"github.com/concourse/concourse/atc"
	"github.com/gobuffalo/packr/v2"
	"github.com/jtarchie/generate-install-pipeline/config"
	"github.com/jtarchie/generate-install-pipeline/resources"
	"sigs.k8s.io/yaml"
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

func (p *Creator) addStepToJob(step atc.Step) {
	p.config.Jobs[0].PlanSequence = append(
		p.config.Jobs[0].PlanSequence,
		step,
	)
}

func (p *Creator) setupSteps() {
	for _, step := range p.payload.Steps {
		p.addStepToJob(step.AsPivnetResource().AsGetStep())

		p.config.Resources = append(
			p.config.Resources,
			step.AsResourceConfig(),
		)
	}
}

func (p *Creator) AsPipeline() atc.Config {
	return p.config
}

func (p *Creator) ensurePutDeployments(step atc.Step) atc.Step {
	return atc.Step{
		Config: &atc.EnsureStep{
			Step: step.Config,
			Hook: atc.Step{
				Config: &atc.PutStep{
					Name: "deployments",
				},
			},
		},
	}
}

//go:generate go run github.com/gobuffalo/packr/v2/packr2
func (p *Creator) setupTasks() error {
	createInfraTask, err := p.getTask("create-infrastructure")
	if err != nil {
		return fmt.Errorf("cannot create infrastructure: %w", err)
	}

	p.addStepToJob(p.ensurePutDeployments(createInfraTask))

	deleteInfraTask, err := p.getTask("delete-infrastructure")
	if err != nil {
		return fmt.Errorf("cannot delete infrastructure: %w", err)
	}

	p.addStepToJob(p.ensurePutDeployments(deleteInfraTask))

	return nil
}

func (p *Creator) getTask(taskName string) (atc.Step, error) {
	box := packr.New("tasks", "./tasks")
	contents, err := box.Find(fmt.Sprintf("%s.yml", taskName))
	if err != nil {
		return atc.Step{}, fmt.Errorf("could not load task %s.yml: %w", taskName, err)
	}

	var taskConfig atc.TaskConfig
	err = yaml.UnmarshalStrict(contents, &taskConfig)

	if err != nil {
		return atc.Step{}, fmt.Errorf("could not unmarshal task %s.yml: %w", taskName, err)
	}

	return atc.Step{
		Config: &atc.TaskStep{
			Name:   taskName,
			Config: &taskConfig,
		},
		UnknownFields: nil,
	}, nil
}

func NewCreator(c config.Payload) (*Creator, error) {
	p := Creator{
		payload: c,
	}
	p.setupResourceTypes()
	p.setupDefaultResources()
	p.setupJobDefaultGets()
	p.setupSteps()

	err := p.setupTasks()
	if err != nil {
		return nil, fmt.Errorf("could not setup tasks: %w", err)
	}

	return &p, nil
}
