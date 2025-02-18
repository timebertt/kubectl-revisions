package e2e

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"

	. "github.com/timebertt/kubectl-revisions/test/e2e/exec"
)

var _ = Describe("version command", func() {
	It("should print the version", func() {
		Eventually(RunPluginAndWait("version")).Should(Say(`kubectl-revisions (\(devel\)|v.+)`))
	})
})
