package main_test

import (
	"io/ioutil"
	"path/filepath"

	"github.com/concourse/concourse/atc"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"sigs.k8s.io/yaml"
)

var _ = Describe("when using the tasks", func() {
	It("has tasks that pass shellcheck", func() {
		tasks, err := filepath.Glob("pipeline/tasks/*.yml")
		Expect(err).NotTo(HaveOccurred())

		Expect(len(tasks)).To(BeNumerically(">", 0))

		for _, taskFilename := range tasks {
			contents, err := ioutil.ReadFile(taskFilename)
			Expect(err).NotTo(HaveOccurred())

			var taskConfig atc.TaskConfig
			err = yaml.UnmarshalStrict(contents, &taskConfig)
			Expect(err).NotTo(HaveOccurred())

			scriptFile, err := ioutil.TempFile("", "")
			Expect(err).NotTo(HaveOccurred())

			Expect(taskConfig.Run.Path).To(Equal("bash"))
			Expect(taskConfig.Run.Args).To(HaveLen(2))

			_, err = scriptFile.WriteString(taskConfig.Run.Args[1])
			Expect(err).NotTo(HaveOccurred())

			err = scriptFile.Close()
			Expect(err).NotTo(HaveOccurred())

			session, _, _ := run("shellcheck", "-s", "bash", scriptFile.Name())
			Eventually(session).Should(gexec.Exit(0))
		}
	})
})
