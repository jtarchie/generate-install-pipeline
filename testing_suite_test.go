package main_test

import (
	"io"
	"io/ioutil"
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf/om/interpolate"
)

func TestTesting(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Testing Suite")
}

//nolint:gochecknoglobals
var binPath string

var _ = BeforeSuite(func() {
	var err error

	binPath, err = gexec.Build("main.go")
	Expect(err).NotTo(HaveOccurred())
})

func run(name string, args ...string) (*gexec.Session, *gbytes.Buffer, *gbytes.Buffer) {
	command := exec.Command(name, args...)

	stdout := gbytes.NewBuffer()
	stderr := gbytes.NewBuffer()

	session, err := gexec.Start(
		command,
		io.MultiWriter(GinkgoWriter, stdout),
		io.MultiWriter(GinkgoWriter, stderr),
	)
	Expect(err).NotTo(HaveOccurred())

	return session, stdout, stderr
}

func writeFile(contents string) string {
	file, err := ioutil.TempFile("", "")
	Expect(err).NotTo(HaveOccurred())

	_, err = file.WriteString(contents)
	Expect(err).NotTo(HaveOccurred())

	err = file.Close()
	Expect(err).NotTo(HaveOccurred())

	return file.Name()
}

func path(stdout *gbytes.Buffer, lookup string) string {
	bytes, err := interpolate.Execute(interpolate.Options{
		TemplateFile: writeFile(string(stdout.Contents())),
		Path:         lookup,
	})
	Expect(err).NotTo(HaveOccurred())

	return string(bytes)
}

var _ = Describe("When providing a configuration", func() {
	When("no config file is provided", func() {
		It("fails with an error message", func() {
			session, stdout, stderr := run(binPath)

			Eventually(session).Should(gexec.Exit(1))
			Expect(string(stdout.Contents())).To(Equal(""))
			Expect(string(stderr.Contents())).To(ContainSubstring("--config is required"))
		})
	})

	When("the step is a tile", func() {
		It("includes the resource", func() {
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
  environments:
  - name: testing
    iaas: gcp
`),
			)

			Eventually(session).Should(gexec.Exit(0))

			By("declaring it in the resources")
			Expect(path(stdout, "/resources/name=tile-elastic-runtime-2.0.0")).To(MatchYAML(`
name: tile-elastic-runtime-2.0.0
source:
  api_token: ((pivnet.api_token))
  product_slug: elastic-runtime
  product_version: 2\.0\.0
type: pivnet
`))

			By("declaring it in the job")
			Expect(path(stdout, "/jobs/name=build-testing/plan/get=tile-elastic-runtime-2.0.0")).To(MatchYAML(`get: tile-elastic-runtime-2.0.0`))
		})
	})

	When("loading default resources", func() {
		It("has paving and platform automation", func() {
			session, stdout, _ := run(
				binPath,
				"--config",
				writeFile(`
deployment:
  uri: "git@github.com:user/repo"
  environments:
  - name: testing
    iaas: gcp
`),
			)

			Eventually(session).Should(gexec.Exit(0))
			Expect(path(stdout, "/resources/name=platform-automation")).To(MatchYAML(`
name: platform-automation
source:
  api_token: ((pivnet.api_token))
  product_slug: platform-automation
  product_version: .*
type: pivnet
`))
			Expect(path(stdout, "/jobs/name=build-testing/plan/get=platform-automation-image")).To(MatchYAML(`
get: platform-automation-image
resource: platform-automation
params:
  globs: ['*image*.tgz']
  unpack: true
`))
			Expect(path(stdout, "/jobs/name=build-testing/plan/get=platform-automation-tasks")).To(MatchYAML(`
get: platform-automation-tasks
resource: platform-automation
params:
  globs: ['*tasks*.zip']
  unpack: true
`))
			Expect(path(stdout, "/resources/name=paving")).To(MatchYAML(`
name: paving
source:
  uri: https://github.com/pivotal/paving
type: git
`))
			Expect(path(stdout, "/jobs/name=build-testing/plan/get=paving")).To(MatchYAML(`get: paving`))
		})
	})

	When("specifying a deployment", func() {
		It("adds a private key when ssh URI", func() {
			session, stdout, _ := run(
				binPath,
				"--config",
				writeFile(`
deployment:
  uri: "git@github.com:user/repo"
  environments:
  - name: testing
    iaas: gcp
`),
			)

			Eventually(session).Should(gexec.Exit(0))

			Expect(path(stdout, "/resources/name=deployments")).To(MatchYAML(`
name: deployments
source:
  private_key: ((deployments.private_key))
  uri: git@github.com:user/repo
type: git
`))
			Expect(path(stdout, "/jobs/name=build-testing/plan/get=deployments")).To(MatchYAML(`get: deployments`))
		})
	})
})
