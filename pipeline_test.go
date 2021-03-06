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
- ops-manager:
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

			Expect(path(stdout, "/jobs/name=build-testing/plan/task=create-infrastructure/config/platform")).To(Equal("linux\n"))
			Expect(path(stdout, "/jobs/name=build-testing/plan")).To(ContainsYAML(`
- get: deployments
- get: paving
- get: platform-automation-image
- get: platform-automation-tasks
- get: ops-manager-2.0.0
- task: create-infrastructure
  params:
    IAAS: gcp
    DEPLOYMENT_NAME: testing
  ensure:
    put: deployments
    params:
      repository: deployments
      rebase: true
- task: create-env
  ensure:
    put: deployments
- task: prepare-tasks-with-secrets
  file: platform-automation-tasks/tasks/prepare-tasks-with-secrets.yml
- task: create-vm-ops-manager-2.0.0
  file: platform-automation-tasks/tasks/create-vm.yml
  ensure:
    task: state-file
- task: configure-authentication
  file: platform-automation-tasks/tasks/configure-authentication.yml
- task: configure-director
  file: platform-automation-tasks/tasks/configure-directory.yml
- task: apply-changes
  file: platform-automation-tasks/tasks/apply-changes.yml
- task: delete-installation
  file: platform-automation-tasks/tasks/delete-installation.yml
- task: delete-vm
- task: delete-infrastructure
  ensure:
    put: deployments
`))
		})
	})
})
