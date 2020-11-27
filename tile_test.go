package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("the step is a tile", func() {
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
  branch: main
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
