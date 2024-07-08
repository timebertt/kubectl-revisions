package workload

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	. "github.com/timebertt/kubectl-revisions/pkg/test/matcher"
)

func PrepareTestNamespace(optionalName ...string) string {
	GinkgoHelper()

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "test-",
			Labels:       CommonLabels(),
		},
	}

	if len(optionalName) > 0 {
		namespace.Name = optionalName[0]
		namespace.GenerateName = ""
	}

	Expect(testClient.Create(context.Background(), namespace)).To(Succeed())
	logf.Log.Info("Created test namespace", "namespace", namespace.Name)

	DeferCleanup(func() {
		DumpNamespaceContents(namespace.Name)
	})

	DeferCleanup(func() {
		logf.Log.Info("Deleting test namespace", "namespace", namespace.Name)
		Expect(testClient.Delete(context.Background(), namespace)).To(Or(Succeed(), BeNotFoundError()))
	})

	return namespace.Name
}

// DumpNamespaceContents dumps all relevant objects in the test namespace to a dedicated ARTIFACTS directory if the
// spec failed to help deflaking/debugging tests.
func DumpNamespaceContents(namespace string) {
	dir := os.Getenv("ARTIFACTS")
	if !CurrentSpecReport().Failed() || dir == "" {
		return
	}

	dir = filepath.Join(dir, namespace)
	logf.Log.Info("Dumping contents of test namespace", "namespace", namespace, "dir", dir)

	// nolint:gosec // this is test code
	Expect(os.MkdirAll(dir, 0755)).To(Succeed())

	DumpEventsInNamespace(namespace, dir)

	for _, kind := range []string{
		"pods",
		"deployments",
		"replicasets",
		"statefulsets",
		"daemonsets",
		"controllerrevisions",
	} {
		DumpObjectsInNamespace(namespace, kind, dir)
	}
}

func DumpEventsInNamespace(namespace, dir string) {
	// nolint:gosec // this is test code
	file, err := os.OpenFile(filepath.Join(dir, "events.log"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	Expect(err).NotTo(HaveOccurred())

	session, err := gexec.Start(exec.Command("kubectl", "-n", namespace, "get", "events"), file, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(0))
}

func DumpObjectsInNamespace(namespace, kind, dir string) {
	// nolint:gosec // this is test code
	file, err := os.OpenFile(filepath.Join(dir, kind+".yaml"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	Expect(err).NotTo(HaveOccurred())

	session, err := gexec.Start(exec.Command("kubectl", "-n", namespace, "get", kind, "-oyaml"), file, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(0))
}
