package main_test

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/concourse/concourse/atc"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"sigs.k8s.io/yaml"
)

func tasks() []atc.TaskStep {
	taskFiles, err := filepath.Glob("pipeline/tasks/*.yml")
	Expect(err).NotTo(HaveOccurred())

	Expect(len(taskFiles)).To(BeNumerically(">", 0))

	tasks := []atc.TaskStep{}

	for _, taskFilename := range taskFiles {
		contents, err := ioutil.ReadFile(taskFilename)
		Expect(err).NotTo(HaveOccurred())

		var taskConfig atc.TaskConfig
		err = yaml.UnmarshalStrict(contents, &taskConfig)
		Expect(err).NotTo(HaveOccurred())

		tasks = append(tasks, atc.TaskStep{
			Name:   strings.TrimSuffix(filepath.Base(taskFilename), ".yml"),
			Config: &taskConfig,
		})
	}

	return tasks
}

var _ = Describe("when using the tasks", func() {
	It("includes the unknown environment variables as params", func() {
		session, stdout, _ := run(
			binPath,
			"--config",
			writeFile(`
steps:
- tile:
    slug: elastic-runtime
    version: 2.0.0
deployment:
  uri: "git@github.com:user/repo"
  branch: main
  environments:
  - name: testing
    iaas: gcp
`),
		)

		Eventually(session).Should(gexec.Exit(0))

		findVars := regexp.MustCompile(`\$[A-Z_]+`)

		for _, task := range tasks() {
			taskConfig := task.Config

			Expect(taskConfig.Run.Args).To(HaveLen(2))

			script := taskConfig.Run.Args[1]
			definedVars := findVars.FindAllString(script, -1)

			Expect(len(definedVars)).To(BeNumerically(">", 0))

			allowList := map[string]struct{}{
				"$PWD": {},
			}

			for _, definedVar := range definedVars {
				if _, ok := allowList[definedVar]; ok {
					continue
				}

				findVar := regexp.MustCompile(fmt.Sprintf("%s=", definedVar))
				if !findVar.MatchString(script) {
					Expect(path(
						stdout,
						fmt.Sprintf(
							"/jobs/name=build-testing/plan/task=%s/params/%s",
							task.Name,
							strings.TrimPrefix(definedVar, "$"),
						),
					)).NotTo(Equal(""))
				}
			}
		}
	})

	It("has tasks that pass shellcheck", func() {
		for _, task := range tasks() {
			taskConfig := task.Config

			scriptFile, err := ioutil.TempFile("", "")
			Expect(err).NotTo(HaveOccurred())

			Expect(taskConfig.Run.Args).To(HaveLen(2))

			shell := taskConfig.Run.Path
			Expect([]string{"sh", "bash"}).To(ContainElement(shell))

			_, err = scriptFile.WriteString(taskConfig.Run.Args[1])
			Expect(err).NotTo(HaveOccurred())

			err = scriptFile.Close()
			Expect(err).NotTo(HaveOccurred())

			session, _, _ := run("shellcheck", "-s", shell, scriptFile.Name())
			Eventually(session).Should(gexec.Exit(0))
		}
	})
})
