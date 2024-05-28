package exec

import (
	"flag"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	binaryPath string
	tmpPath    string
)

func AddFlags() {
	flag.StringVar(&binaryPath, "binary-path", "", "Specify a pre-built binary to test instead of building a binary during test execution")
}

func PrepareTestBinaries() {
	if binaryPath != "" {
		logf.Log.Info("Using pre-built binary", "path", binaryPath)
	} else {
		By("Building kubectl-revisions binary")
		var err error
		binaryPath, err = gexec.Build("../..")
		Expect(err).NotTo(HaveOccurred())
		binaryPath = filepath.Join(binaryPath, "kubectl-revisions")
		logf.Log.Info("Using binary", "path", binaryPath)

		DeferCleanup(func() {
			gexec.CleanupBuildArtifacts()
		})
	}

	preparePath()
}

func preparePath() {
	By("Preparing test PATH")
	var err error
	tmpPath, err = os.MkdirTemp("", "e2e-kubectl-revisions-")
	Expect(err).NotTo(HaveOccurred())
	logf.Log.Info("Using tmp dir as PATH", "dir", tmpPath)

	DeferCleanup(func() {
		Expect(os.RemoveAll(tmpPath)).To(Succeed())
	})

	// Symlink all needed binaries to a dedicated PATH dir.
	// Use a single-dir PATH with only the needed binaries inside instead of appending the user's PATH to run against a
	// clean test environment.

	// symlink the kubectl-revisions binary so that kubectl can find the plugin
	absoluteBinaryPath, err := filepath.Abs(binaryPath)
	Expect(err).NotTo(HaveOccurred())
	Expect(os.Symlink(absoluteBinaryPath, filepath.Join(tmpPath, "kubectl-revisions"))).To(Succeed())

	// diff is used as an external diff command
	diffPath, err := exec.LookPath("diff")
	Expect(err).NotTo(HaveOccurred(), "diff is required in PATH for e2e tests")
	Expect(os.Symlink(diffPath, filepath.Join(tmpPath, "diff"))).To(Succeed())

	// cat is used as an alternative external "diff" command
	catPath, err := exec.LookPath("cat")
	Expect(err).NotTo(HaveOccurred(), "cat is required in PATH for e2e tests")
	Expect(os.Symlink(catPath, filepath.Join(tmpPath, "cat"))).To(Succeed())
}

func NewHistoryCommand(args ...string) *exec.Cmd {
	// nolint:gosec // no security risk in shared test code
	command := exec.Command("kubectl", append([]string{"history"}, args...)...)
	command.Env = append(command.Environ(), "PATH="+tmpPath)

	return command
}

func RunCommand(cmd *exec.Cmd) *gexec.Session {
	GinkgoHelper()

	session, err := gexec.Start(
		cmd,
		gexec.NewPrefixedWriter("[out] ", GinkgoWriter),
		gexec.NewPrefixedWriter("[err] ", GinkgoWriter),
	)
	Expect(err).NotTo(HaveOccurred())

	return session
}

func Wait(session *gexec.Session) *gexec.Session {
	GinkgoHelper()

	Eventually(session).Should(gexec.Exit(0))
	return session
}

func RunHistory(args ...string) *gexec.Session {
	return RunCommand(NewHistoryCommand(args...))
}

func RunHistoryAndWait(args ...string) *gexec.Session {
	return Wait(RunHistory(args...))
}
