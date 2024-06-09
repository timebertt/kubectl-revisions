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

		Eventually(session).Should(Say(`Synopsis:\n`))
		Eventually(session).Should(Say(`Available Commands:\n`))
		Eventually(session).Should(Say(`\s+get\s+`))
		Eventually(session).Should(Say(`\s+diff\s+`))
		Eventually(session).Should(Say(`Other Commands:\n`))
		Eventually(session).Should(Say(`\s+completion\s+`))
		Eventually(session).Should(Say(`\s+version\s+`))
		Eventually(session).Should(Say(`Help Commands:\n`))
		Eventually(session).Should(Say(`\s+options\s+`))
		Eventually(session).Should(Say(`\s+help\s+`))
		Eventually(session).Should(Say(`Usage:`))
		Eventually(session).Should(Say(`kubectl revisions \[command\]\n`))
	}

	Describe("root command", func() {
		It("should print help without args", func() {
			expectHelp(RunPluginAndWait())
		})

		It("should print help on -h arg", func() {
			expectHelp(RunPluginAndWait("-h"))
		})

		It("should print help on --help arg", func() {
			expectHelp(RunPluginAndWait("--help"))
		})
	})

	Describe("help command", func() {
		It("should print help", func() {
			expectHelp(RunPluginAndWait("help"))
		})
	})

	Describe("options command", func() {
		It("should print global flags usage", func() {
			session := RunPluginAndWait("options")

			Eventually(session).Should(Say(`\s+--cluster='':\n`))
			Eventually(session).Should(Say(`\s+-n, --namespace='':\n`))
		})
	})
})
