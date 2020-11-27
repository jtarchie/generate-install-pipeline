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
		Version:  "*",
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
		Branch:   p.payload.Deployment.Branch,
	}
}

func (p Creator) pavingResource() resources.GitResource {
	return resources.GitResource{
		Resource: resources.Resource{Name: "paving"},
		URI:      "https://github.com/pivotal/paving",
	}
}

func (p *Creator) setupJobDefaultGets(env config.Environment) {
	p.config.Jobs = append(
		p.config.Jobs,
		atc.JobConfig{
			Name:   fmt.Sprintf("build-%s", env.Name),
			Serial: true,
		},
	)

	p.addStepToJob(p.deploymentResource().AsGetStep())
	p.addStepToJob(p.pavingResource().AsGetStep())
	p.addStepToJob(p.platformAutomationImageResource().AsGetStep())
	p.addStepToJob(p.platformAutomationTasksResource().AsGetStep())
}

func (p *Creator) lastJobIndex() int {
	return len(p.config.Jobs) - 1
}

func (p *Creator) addStepToJob(step atc.Step) {
	p.config.Jobs[p.lastJobIndex()].PlanSequence = append(
		p.config.Jobs[p.lastJobIndex()].PlanSequence,
		step,
	)
}

func (p *Creator) setupSteps(env config.Environment) {
	for _, step := range p.payload.Steps {
		p.addStepToJob(step.AsPivnetResource(env).AsGetStep())

		p.config.Resources = append(
			p.config.Resources,
			step.AsPivnetResource(env).AsResourceConfig(),
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
					Params: map[string]interface{}{
						"repository": "deployments",
						"rebase":     true,
					},
				},
			},
		},
	}
}

//go:generate go run github.com/gobuffalo/packr/v2/packr2
func (p *Creator) setupTasks(env config.Environment) error {
	createInfraTask, err := p.getTask("create-infrastructure", map[string]string{
		"DEPLOYMENT_NAME": env.Name,
		"IAAS":            env.IAAS,
	}, "")
	if err != nil {
		return fmt.Errorf("cannot create infrastructure: %w", err)
	}

	p.addStepToJob(p.ensurePutDeployments(createInfraTask))

	createEnvTask, err := p.getTask("create-env", map[string]string{
		"DEPLOYMENT_NAME": env.Name,
		"OM_USERNAME":     "((om.username))",
		"OM_PASSWORD":     "((om.password))",
	}, "platform-automation-image")
	if err != nil {
		return fmt.Errorf("cannot create env: %w", err)
	}

	p.addStepToJob(p.ensurePutDeployments(createEnvTask))
	p.addStepToJob(p.addPATask("prepare-tasks-with-secrets"))

	deleteInfraTask, err := p.getTask("delete-infrastructure", map[string]string{
		"DEPLOYMENT_NAME": env.Name,
		"IAAS":            env.IAAS,
	}, "")
	if err != nil {
		return fmt.Errorf("cannot delete infrastructure: %w", err)
	}

	p.addStepToJob(p.ensurePutDeployments(deleteInfraTask))

	return nil
}

func (p *Creator) getTask(
	taskName string,
	params map[string]string,
	imageName string,
) (atc.Step, error) {
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

	task := &atc.TaskStep{
		Name:   taskName,
		Config: &taskConfig,
		Params: params,
	}

	if imageName != "" {
		task.ImageArtifactName = imageName
	}

	return atc.Step{
		Config: task,
	}, nil
}

func (p *Creator) addPATask(taskName string) atc.Step {
	task := &atc.TaskStep{
		Name:              taskName,
		ImageArtifactName: "platform-automation-image",
		ConfigPath:        fmt.Sprintf("platform-automation-tasks/tasks/%s.yml", taskName),
	}

	return atc.Step{
		Config:        task,
	}
}

func NewCreator(c config.Payload) (*Creator, error) {
	err := validate(c)
	if err != nil {
		return nil, fmt.Errorf("config file incomplete: %w", err)
	}

	p := Creator{
		payload: c,
	}
	p.setupResourceTypes()
	p.setupDefaultResources()

	for _, env := range c.Deployment.Environments {
		p.setupJobDefaultGets(env)
		p.setupSteps(env)

		err = p.setupTasks(env)
		if err != nil {
			return nil, fmt.Errorf("could not setup tasks: %w", err)
		}
	}

	return &p, nil
}

func validate(c config.Payload) error {
	if len(c.Deployment.Environments) == 0 {
		return fmt.Errorf("at least one environment is required from deployments.environments[]")
	}

	if c.Deployment.URI == "" {
		return fmt.Errorf("a uri is required to read/write deployments.uri")
	}

	if c.Deployment.Branch == "" {
		return fmt.Errorf("a branch is required to read/write deployments.branch")
	}

	for index, env := range c.Deployment.Environments {
		switch env.IAAS {
		case "openstack", "gcp", "aws", "azure", "vsphere":
			continue
		default:
			return fmt.Errorf("iaas %q unsupported in the deployment.environments[%d]", env.IAAS, index)
		}
	}

	return nil
}
