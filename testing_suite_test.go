package main_test

import (
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf/om/interpolate"
	"io"
	"io/ioutil"
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTesting(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Testing Suite")
}

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

func path(stdout *gbytes.Buffer, lookup string) []byte {
	bytes, err := interpolate.Execute(interpolate.Options{
		TemplateFile: writeFile(string(stdout.Contents())),
		Path:         lookup,
	})
	Expect(err).NotTo(HaveOccurred())

	return bytes
}

var _ = Describe("When providing a configuration", func() {
	When("the step is OpsManager", func() {
		It("includes the resource", func() {
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

			By("declaring it in the resources")
			Expect(path(stdout, "/resources/name=opsmanager-2.0.0")).To(MatchYAML(`
name: opsmanager-2.0.0
type: pivnet
source:
  api_token: ((pivnet.api_token))
  product_slug: opsmanager
  product_version: 2\.0\.0
`))

			By("declaring it in the job")
			Expect(path(stdout, "/jobs/name=build/plan/get=opsmanager-2.0.0")).To(MatchYAML(`get: opsmanager-2.0.0`))
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
			Expect(path(stdout, "/jobs/name=build/plan/get=tile-elastic-runtime-2.0.0")).To(MatchYAML(`get: tile-elastic-runtime-2.0.0`))
		})
	})
})
