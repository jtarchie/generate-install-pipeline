package main

import (
	"flag"
	"fmt"
	"github.com/concourse/concourse/atc"
	"github.com/concourse/concourse/atc/configvalidate"
	"io/ioutil"
	"log"
	"os"
	"sigs.k8s.io/yaml"
	"strings"
)

func main() {
	err := execute()
	if err != nil {
		log.Fatalf("main: %s", err)
	}
}

func execute() error {
	configFilename := flag.String("config", "", "config file that describes the pipeline to build")
	flag.Parse()

	payload := Payload{}

	err := parseYAML(*configFilename, &payload)
	if err != nil {
		return fmt.Errorf("could not read payload: %w", err)
	}

	var config atc.Config
	config.ResourceTypes = append(
		config.ResourceTypes,
		atc.ResourceType{
			Name: "pivnet",
			Type: "registry-image",
			Source: map[string]interface{}{
				"repository": "pivotalcf/pivnet-resource",
				"tag":        "latest-final",
			},
		},
	)
	pavingResource := GitResource{
		Resource: Resource{Name: "paving"},
		URI:      "https://github.com/pivotal/paving",
	}
	deploymentsResource := GitResource{
		Resource: Resource{Name: "deployments"},
		URI:      "https://github.com/pivotal/paving",
	}
	ciResource := GitResource{
		Resource: Resource{Name: "docs-platform-automation"},
		URI:      "https://github.com/pivotal/docs-platform-automation",
	}
	platformAutomationResource := PivnetResource{
		Resource: Resource{Name: "platform-automation"},
		Slug:     "platform-automation",
		Version:  ".*",
	}

	config.Resources = append(
		config.Resources,
		ciResource.AsResourceConfig(),
		deploymentsResource.AsResourceConfig(),
		pavingResource.AsResourceConfig(),
		platformAutomationResource.AsResourceConfig(),
	)

	jobConfig := atc.JobConfig{
		Name:         "build",
		Serial:       true,
		PlanSequence: nil,
	}

	jobConfig.PlanSequence = append(jobConfig.PlanSequence,
		ciResource.AsGetStep(),
		deploymentsResource.AsGetStep(),
		pavingResource.AsGetStep(),
		platformAutomationResource.AsGetStep(),
	)

	for _, step := range payload.Steps {
		jobConfig.PlanSequence = append(
			jobConfig.PlanSequence,
			step.AsPivnetResource().AsGetStep(),
		)
		config.Resources = append(
			config.Resources,
			step.AsResourceConfig(),
		)
	}

	config.Jobs = append(
		config.Jobs,
		jobConfig,
	)

	contents, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("could not generate pipeline YAML: %w", err)
	}

	fmt.Printf("%s\n", contents)

	warnings, errors := configvalidate.Validate(config)
	if len(warnings) > 0 {
		_, _ = fmt.Fprintln(os.Stderr, "Warnings: ")
		for _, msg := range warnings {
			_, _ = fmt.Fprintf(os.Stderr, "* %s\n", msg)
		}
	}
	if len(errors) > 0 {
		_, _ = fmt.Fprintln(os.Stderr, "Errors: ")
		for _, msg := range errors {
			_, _ = fmt.Fprintf(os.Stderr, "* %s\n", msg)
		}

		return fmt.Errorf("there are errors with the generated pipeline")
	}

	return nil
}

func parseYAML(filename string, payload interface{}) error {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("could not read %q: %w", filename, err)
	}

	err = yaml.UnmarshalStrict(contents, payload)
	if err != nil {
		return fmt.Errorf("could not unmarshal contents of %q: %w", filename, err)
	}

	return nil
}

type Deployment struct {
	Name string `yaml:"name"`
	IAAS string `yaml:"iaas"`
}
type Deployments []Deployment

type StepOpsManager struct {
	Version string `yaml:"version"`
}
type StepTile struct {
	Slug    string `yaml:"slug"`
	Version string `yaml:"version"`
}
type Step struct {
	OpsManager *StepOpsManager `yaml:"opsmanager"`
	Tile       *StepTile       `yaml:"tile"`
}

func (s Step) ResourceName() string {
	name := ""
	switch {
	case s.OpsManager != nil:
		name = fmt.Sprintf("opsmanager-%s", s.OpsManager.Version)
	case s.Tile != nil:
		name = fmt.Sprintf("tile-%s-%s", s.Tile.Slug, s.Tile.Version)
	}

	name = strings.ReplaceAll(name, "*", "x")
	return name
}

func (s Step) AsPivnetResource() PivnetResource {
	pivnetResource := PivnetResource{}
	pivnetResource.Name = s.ResourceName()

	switch {
	case s.OpsManager != nil:
		pivnetResource.Slug = "opsmanager"
		pivnetResource.Version = s.OpsManager.Version
	case s.Tile != nil:
		pivnetResource.Slug = s.Tile.Slug
		pivnetResource.Version = s.Tile.Version
	}

	return pivnetResource
}

func (s Step) AsResourceConfig() atc.ResourceConfig {
	return s.AsPivnetResource().AsResourceConfig()
}

type Steps []Step

type Payload struct {
	Steps       Steps       `yaml:"steps"`
	Deployments Deployments `yaml:"deployments"`
}

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
