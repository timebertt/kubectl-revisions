package e2e

import (
	"flag"
	"os"
	"testing"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/envtest/komega"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/timebertt/kubectl-revisions/test/e2e/exec"
	"github.com/timebertt/kubectl-revisions/test/e2e/workload"
)

func TestMain(m *testing.M) {
	addFlags()
	flag.Parse()

	os.Exit(m.Run())
}

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "kubectl-revisions E2E Test Suite")
}

var (
	log        logr.Logger
	testClient client.Client
	decoder    runtime.Decoder
)

func addFlags() {
	exec.AddFlags()
}

var _ = BeforeSuite(func() {
	SetDefaultEventuallyPollingInterval(100 * time.Millisecond)
	SetDefaultEventuallyTimeout(15 * time.Minute)
	SetDefaultConsistentlyPollingInterval(500 * time.Millisecond)
	SetDefaultConsistentlyDuration(5 * time.Second)

	log = zap.New(zap.UseDevMode(true), zap.WriteTo(GinkgoWriter))
	logf.SetLogger(log)

	restConfig, err := config.GetConfig()
	Expect(err).NotTo(HaveOccurred())

	scheme := runtime.NewScheme()
	schemeBuilder := runtime.NewSchemeBuilder(
		clientgoscheme.AddToScheme,
	)
	Expect(schemeBuilder.AddToScheme(scheme)).To(Succeed())
	decoder = serializer.NewCodecFactory(scheme).UniversalDeserializer()

	testClient, err = client.New(restConfig, client.Options{Scheme: scheme})
	Expect(err).NotTo(HaveOccurred())

	komega.SetClient(testClient)
	workload.SetClient(testClient)

	exec.PrepareTestBinaries()
})
