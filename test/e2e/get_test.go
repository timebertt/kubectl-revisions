package e2e

import (
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/timebertt/kubectl-history/test/e2e/exec"
	"github.com/timebertt/kubectl-history/test/e2e/workload"
)

var _ = Describe("get command", func() {
	var (
		namespace string
		object    client.Object

		args []string
	)

	BeforeEach(func() {
		namespace = workload.PrepareTestNamespace()
		args = []string{"get", "-n", namespace}
	})

	Describe("command aliases", func() {
		BeforeEach(func() {
			object = workload.CreateDeployment(namespace)
			args = append(args, "deployment", object.GetName())
		})

		It("should work with alias ls", func() {
			args[0] = "ls"
			Eventually(RunHistoryAndWait(args...)).Should(Say(`nginx-\S+\s+1\s+\S+\n`))
		})

		It("should work with alias list", func() {
			args[0] = "list"
			Eventually(RunHistoryAndWait(args...)).Should(Say(`nginx-\S+\s+1\s+\S+\n`))
		})
	})

	testCommon := func() {
		It("should print a single revision in list format", func() {
			session := RunHistoryAndWait(args...)
			Eventually(session).Should(Say(`nginx-\S+\s+1\s+\S+\n`))
			Consistently(session).ShouldNot(Say(`nginx-`))
		})

		It("should print image column in wide format", func() {
			workload.BumpImage(object)
			workload.BumpImage(object)

			session := RunHistoryAndWait(append(args, "-o", "wide")...)
			Eventually(session).Should(Say(`nginx-\S+\s+1\s+\S+\s+nginx\s+\S+:0.1\n`))
			Eventually(session).Should(Say(`nginx-\S+\s+2\s+\S+\s+nginx\s+\S+:0.2\n`))
			Eventually(session).Should(Say(`nginx-\S+\s+3\s+\S+\s+nginx\s+\S+:0.3\n`))
			Consistently(session).ShouldNot(Say(`nginx-`))
		})

		It("should print a specific revision in list format (absolute revision)", func() {
			workload.BumpImage(object)
			workload.BumpImage(object)

			session := RunHistoryAndWait(append(args, "--revision=2")...)
			Eventually(session).Should(Say(`nginx-\S+\s+2\s+\S+\n`))
			Consistently(session).ShouldNot(Say(`nginx-`))
		})

		It("should print a specific revision in wide format (relative revision)", func() {
			workload.BumpImage(object)
			workload.BumpImage(object)

			session := RunHistoryAndWait(append(args, "--revision=2", "-o", "wide")...)
			Eventually(session).Should(Say(`nginx-\S+\s+2\s+\S+\s+nginx\s+\S+:0.2\n`))
			Consistently(session).ShouldNot(Say(`nginx-`))
		})

		It("should print a specific revision in yaml format", func() {
			workload.BumpImage(object)
			workload.BumpImage(object)

			session := RunHistoryAndWait(append(args, "--revision=1", "-o", "yaml")...)

			yamlBytes, err := io.ReadAll(session.Out)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(yamlBytes)).To(ContainSubstring("image: " + workload.ImageRepository + ":0.1"))

			Expect(runtime.DecodeInto(decoder, yamlBytes, workload.RevisionObjectFor(object))).To(Succeed())
		})

		It("should print a specific revision's pod template in json format on --template-only", func() {
			workload.BumpImage(object)
			workload.BumpImage(object)

			session := RunHistoryAndWait(append(args, "--revision=2", "-o", "json", "--template-only")...)

			jsonBytes, err := io.ReadAll(session.Out)
			Expect(err).NotTo(HaveOccurred())

			pod := &corev1.Pod{}
			Expect(runtime.DecodeInto(decoder, jsonBytes, pod)).To(Succeed())
			Expect(pod.Spec.Containers).To(HaveLen(1))
			Expect(pod.Spec.Containers[0].Image).To(Equal(workload.ImageRepository + ":0.2"))
		})

		It("should print the full revision list in yaml format", func() {
			workload.BumpImage(object)
			workload.BumpImage(object)

			session := RunHistoryAndWait(append(args, "-o", "yaml")...)

			yamlBytes, err := io.ReadAll(session.Out)
			Expect(err).NotTo(HaveOccurred())

			list := &metav1.List{}
			Expect(runtime.DecodeInto(decoder, yamlBytes, list)).To(Succeed())
			Expect(list.Items).To(HaveLen(3))

			Expect(runtime.DecodeInto(decoder, list.Items[0].Raw, workload.RevisionObjectFor(object))).To(Succeed())
		})
	}

	Context("Deployment", func() {
		BeforeEach(func() {
			object = workload.CreateDeployment(namespace)
			args = append(args, "deployment", object.GetName())
		})

		testCommon()

		It("should work with short type", func() {
			args[3] = "deploy"
			Eventually(RunHistoryAndWait(args...)).Should(Say(`nginx-\S+\s+1\s+\S+\n`))
		})

		It("should work with grouped type", func() {
			args[3] = "deployments.apps"
			Eventually(RunHistoryAndWait(args...)).Should(Say(`nginx-\S+\s+1\s+\S+\n`))
		})

		It("should work with fully-qualified type", func() {
			args[3] = "deployments.v1.apps"
			Eventually(RunHistoryAndWait(args...)).Should(Say(`nginx-\S+\s+1\s+\S+\n`))
		})

		It("should work with slash name", func() {
			args[3] = "deployment/nginx"
			args = args[:len(args)-1]
			Eventually(RunHistoryAndWait(args...)).Should(Say(`nginx-\S+\s+1\s+\S+\n`))
		})
	})

	Context("StatefulSet", func() {
		BeforeEach(func() {
			object = workload.CreateStatefulSet(namespace)
			args = append(args, "statefulset", object.GetName())
		})

		testCommon()
	})

	Context("DaemonSet", func() {
		BeforeEach(func() {
			object = workload.CreateDaemonSet(namespace)
			args = append(args, "daemonset", object.GetName())
		})

		testCommon()
	})
})
