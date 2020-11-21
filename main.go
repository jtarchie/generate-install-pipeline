package main

import (
	"flag"
	"fmt"
	"github.com/concourse/concourse/atc"
	"github.com/concourse/concourse/atc/configvalidate"
	"github.com/jtarchie/generate-install-pipeline/config"
	"github.com/jtarchie/generate-install-pipeline/resources"
	"io/ioutil"
	"log"
	"os"
	"sigs.k8s.io/yaml"
)

func main() {
	err := execute()
	if err != nil {
		log.Fatalf("main: %s", err)
	}
}

func execute() error {
	configFilename := flag.String("config", "", "pipeline file that describes the pipeline to build")
	flag.Parse()

	if *configFilename == "" {
		return fmt.Errorf("--config is required")
	}

	payload := config.Payload{}

	err := parseYAML(*configFilename, &payload)
	if err != nil {
		return fmt.Errorf("could not read payload: %w", err)
	}

	var pipeline atc.Config
	pipeline.ResourceTypes = append(
		pipeline.ResourceTypes,
		atc.ResourceType{
			Name: "pivnet",
			Type: "registry-image",
			Source: map[string]interface{}{
				"repository": "pivotalcf/pivnet-resource",
				"tag":        "latest-final",
			},
		},
	)
	pavingResource := resources.GitResource{
		Resource: resources.Resource{Name: "paving"},
		URI:      "https://github.com/pivotal/paving",
	}
	deploymentsResource := resources.GitResource{
		Resource: resources.Resource{Name: "deployments"},
		URI:      payload.Deployment.URI,
	}
	platformAutomationResource := resources.PivnetResource{
		Resource: resources.Resource{Name: "platform-automation-image"},
		Slug:     "platform-automation",
		Version:  ".*",
		Globs:    []string{"*image*.tgz"},
	}

	pipeline.Resources = append(
		pipeline.Resources,
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
		deploymentsResource.AsGetStep(),
		pavingResource.AsGetStep(),
		platformAutomationResource.AsGetStep(),
	)

	for _, step := range payload.Steps {
		jobConfig.PlanSequence = append(
			jobConfig.PlanSequence,
			step.AsPivnetResource().AsGetStep(),
		)
		pipeline.Resources = append(
			pipeline.Resources,
			step.AsResourceConfig(),
		)
	}

	pipeline.Jobs = append(
		pipeline.Jobs,
		jobConfig,
	)

	contents, err := yaml.Marshal(pipeline)
	if err != nil {
		return fmt.Errorf("could not generate pipeline YAML: %w", err)
	}

	fmt.Printf("%s\n", contents)

	warnings, errors := configvalidate.Validate(pipeline)
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
