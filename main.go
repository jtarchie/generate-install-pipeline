package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/concourse/concourse/atc"
	"github.com/concourse/concourse/atc/configvalidate"
	"github.com/jtarchie/generate-install-pipeline/config"
	"github.com/jtarchie/generate-install-pipeline/pipeline"
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

	creator := pipeline.NewCreator(payload)
	pipeline := creator.AsPipeline()

	contents, err := yaml.Marshal(pipeline)
	if err != nil {
		return fmt.Errorf("could not generate pipeline YAML: %w", err)
	}

	fmt.Printf("%s\n", contents)

	return displayPipelineWarningsAndErrors(pipeline)
}

func displayPipelineWarningsAndErrors(pipeline atc.Config) error {
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
