package main_test

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("using the examples", func() {
	It("ensures they are all valid", func() {
		examples, err := filepath.Glob("examples/*.yml")
		Expect(err).NotTo(HaveOccurred())

		Expect(len(examples)).To(BeNumerically(">", 0))

		for _, example := range examples {
			session, _, _ := run(
				binPath,
				"--config", example,
			)
			Eventually(session).Should(gexec.Exit(0))
		}
	})
})
