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

	// symlink external binaries needed by e2e tests
	for _, externalBinary := range []string{
		"diff", // diff is used as the default external diff command
		"ls",   // ls is used as an alternative external "diff" command
	} {
		binPath, err := exec.LookPath(externalBinary)
		Expect(err).NotTo(HaveOccurred(), "%s is required in PATH for e2e tests", externalBinary)
		Expect(os.Symlink(binPath, filepath.Join(tmpPath, externalBinary))).To(Succeed())
	}
}

func NewPluginCommand(args ...string) *exec.Cmd {
	// nolint:gosec // no security risk in shared test code
	command := exec.Command("kubectl", append([]string{"revisions"}, args...)...)
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

func RunPlugin(args ...string) *gexec.Session {
	return RunCommand(NewPluginCommand(args...))
}

func RunPluginAndWait(args ...string) *gexec.Session {
	return Wait(RunPlugin(args...))
}
