package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("When creating pipelines", func() {
	When("creating an OpsManager", func() {
		It("includes one creates infrastructure, create-vm, configures, apply changes, and cleanups", func() {
			session, stdout, _ := run(
				binPath,
				"--config",
				writeFile(`
steps:
- opsmanager:
    version: 2.0.0
`),
			)

			Eventually(session).Should(gexec.Exit(0))

			Expect(path(stdout, "/jobs/name=build/plan/task=create-infrastructure/config/platform")).To(Equal("linux\n"))
			Expect(path(stdout, "/jobs/name=build/plan")).To(ContainsYAML(`
- get: deployments
- get: paving
- get: platform-automation-image
- get: platform-automation-tasks
- get: opsmanager-2.0.0
- task: create-infrastructure
  ensure:
    put: deployments
- task: delete-infrastructure
  ensure:
    put: deployments
`))
		})
	})
})
