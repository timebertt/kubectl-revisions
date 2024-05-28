package e2e

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "github.com/timebertt/kubectl-revisions/test/e2e/exec"
)

var _ = Describe("command help", func() {
	expectHelp := func(session *gexec.Session) {
		GinkgoHelper()

		Eventually(session).Should(Say(`Usage:`))
		Eventually(session).Should(Say(`kubectl history \[command\]\n`))
		Eventually(session).Should(Say(`Available Commands:\n`))
		Eventually(session).Should(Say(`\s+diff\s+`))
		Eventually(session).Should(Say(`Other Commands:\n`))
		Eventually(session).Should(Say(`\s+help\s+`))
		Eventually(session).Should(Say(`\s+version\s+`))
	}

	Describe("root command", func() {
		It("should print help without args", func() {
			expectHelp(RunHistoryAndWait())
		})

		It("should print help on -h arg", func() {
			expectHelp(RunHistoryAndWait("-h"))
		})

		It("should print help on --help arg", func() {
			expectHelp(RunHistoryAndWait("--help"))
		})
	})

	Describe("help command", func() {
		It("should print help", func() {
			expectHelp(RunHistoryAndWait("help"))
		})
	})
})
