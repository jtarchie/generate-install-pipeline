package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("When using ops-manager", func() {
	It("includes the resource", func() {
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

		By("declaring it in the resources")
		Expect(path(stdout, "/resources/name=ops-manager-2.0.0")).To(MatchYAML(`
name: ops-manager-2.0.0
type: pivnet
source:
  api_token: ((pivnet.api_token))
  product_slug: ops-manager
  product_version: 2\.0\.0
`))

		By("declaring it in the job")
		Expect(path(stdout, "/jobs/name=build-testing/plan/get=ops-manager-2.0.0")).To(MatchYAML(`
get: ops-manager-2.0.0
params:
  globs: ["*gcp*.yml"]
`))
	})

	It("allows a version wildcard", func() {
		session, stdout, _ := run(
			binPath,
			"--config",
			writeFile(`
steps:
- ops-manager:
    version: 2.0.*
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
		Expect(path(stdout, "/resources/name=ops-manager-2.0.x")).To(MatchYAML(`
name: ops-manager-2.0.x
type: pivnet
source:
  api_token: ((pivnet.api_token))
  product_slug: ops-manager
  product_version: 2\.0\..*
`))

		By("declaring it in the job")
		Expect(path(stdout, "/jobs/name=build-testing/plan/get=ops-manager-2.0.x")).To(MatchYAML(`
get: ops-manager-2.0.x
params:
  globs: ["*gcp*.yml"]
`))
	})

	When("specifying the IAAS as GCP", func() {
		It("includes the GCP glob", func() {
			session, stdout, _ := run(
				binPath,
				"--config",
				writeFile(`
steps:
- ops-manager:
    version: 2.0.*
deployment:
  uri: "git@github.com:user/repo"
  branch: main
  environments:
  - name: testing
    iaas: gcp
`),
			)
			Eventually(session).Should(gexec.Exit(0))

			Expect(path(stdout, "/jobs/name=build-testing/plan/get=ops-manager-2.0.x")).To(MatchYAML(`
get: ops-manager-2.0.x
params:
  globs: ["*gcp*.yml"]
`))
		})
	})

	When("specifying the IAAS as AWS", func() {
		It("includes the AWS glob", func() {
			session, stdout, _ := run(
				binPath,
				"--config",
				writeFile(`
steps:
- ops-manager:
    version: 2.0.*
deployment:
  uri: "git@github.com:user/repo"
  branch: main
  environments:
  - name: testing
    iaas: aws
`),
			)
			Eventually(session).Should(gexec.Exit(0))

			Expect(path(stdout, "/jobs/name=build-testing/plan/get=ops-manager-2.0.x")).To(MatchYAML(`
get: ops-manager-2.0.x
params:
  globs: ["*aws*.yml"]
`))
		})
	})

	When("specifying the IAAS as Azure", func() {
		It("includes the Azure glob", func() {
			session, stdout, _ := run(
				binPath,
				"--config",
				writeFile(`
steps:
- ops-manager:
    version: 2.0.*
deployment:
  uri: "git@github.com:user/repo"
  branch: main
  environments:
  - name: testing
    iaas: azure
`),
			)
			Eventually(session).Should(gexec.Exit(0))

			Expect(path(stdout, "/jobs/name=build-testing/plan/get=ops-manager-2.0.x")).To(MatchYAML(`
get: ops-manager-2.0.x
params:
  globs: ["*azure*.yml"]
`))
		})
	})

	When("specifying the IAAS as vSphere", func() {
		It("includes the vSphere glob", func() {
			session, stdout, _ := run(
				binPath,
				"--config",
				writeFile(`
steps:
- ops-manager:
    version: 2.0.*
deployment:
  uri: "git@github.com:user/repo"
  branch: main
  environments:
  - name: testing
    iaas: vsphere
`),
			)
			Eventually(session).Should(gexec.Exit(0))

			Expect(path(stdout, "/jobs/name=build-testing/plan/get=ops-manager-2.0.x")).To(MatchYAML(`
get: ops-manager-2.0.x
params:
  globs: ["*vsphere*.ova"]
`))
		})
	})

	When("specifying the IAAS as Openstack", func() {
		It("includes the Openstack glob", func() {
			session, stdout, _ := run(
				binPath,
				"--config",
				writeFile(`
steps:
- ops-manager:
    version: 2.0.*
deployment:
  uri: "git@github.com:user/repo"
  branch: main
  environments:
  - name: testing
    iaas: openstack
`),
			)
			Eventually(session).Should(gexec.Exit(0))

			Expect(path(stdout, "/jobs/name=build-testing/plan/get=ops-manager-2.0.x")).To(MatchYAML(`
get: ops-manager-2.0.x
params:
  globs: ["*openstack*.raw"]
`))
		})
	})

	When("specifying an unknown IAAS", func() {
		It("returns an error", func() {
			session, _, stderr := run(
				binPath,
				"--config",
				writeFile(`
steps:
- ops-manager:
    version: 2.0.*
deployment:
  uri: "git@github.com:user/repo"
  branch: main
  environments:
  - name: testing
    iaas: unknown
`),
			)
			Eventually(session).Should(gexec.Exit(1))
			Expect(stderr).To(gbytes.Say(`iaas "unknown" unsupported`))
		})
	})
})
