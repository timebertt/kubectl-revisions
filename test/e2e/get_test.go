package e2e

import (
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest/komega"

	. "github.com/timebertt/kubectl-revisions/test/e2e/exec"
	"github.com/timebertt/kubectl-revisions/test/e2e/workload"
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
			object = workload.CreateDeployment(namespace, workload.AppName)
			args = append(args, "deployment", object.GetName())
		})

		It("should work with alias ls", func() {
			args[0] = "ls"
			Eventually(RunPluginAndWait(args...)).Should(Say(`pause-\S+\s+1\s+\d/\d\s+\S+\n`))
		})

		It("should work with alias list", func() {
			args[0] = "list"
			Eventually(RunPluginAndWait(args...)).Should(Say(`pause-\S+\s+1\s+\d/\d\s+\S+\n`))
		})
	})

	testCommon := func(createObject func(namespace, name string) client.Object) {
		It("should print a single revision in list format", func() {
			session := RunPluginAndWait(args...)
			Eventually(session).Should(Say(`NAME\s+REVISION\s+READY\s+AGE\n`))
			Eventually(session).Should(Say(`pause-\S+\s+1\s+\d/\d\s+\S+\n`))
			Consistently(session).ShouldNot(Say(`pause-`))
		})

		It("should print image column in wide format", func() {
			workload.BumpImage(object)
			workload.BumpImage(object)

			session := RunPluginAndWait(append(args, "-o", "wide")...)
			Eventually(session).Should(Say(`NAME\s+REVISION\s+READY\s+AGE\s+CONTAINERS\s+IMAGES\n`))
			Eventually(session).Should(Say(`pause-\S+\s+1\s+\d/\d\s+\S+\s+pause\s+\S+:0.1\n`))
			Eventually(session).Should(Say(`pause-\S+\s+2\s+\d/\d\s+\S+\s+pause\s+\S+:0.2\n`))
			Eventually(session).Should(Say(`pause-\S+\s+3\s+\d/\d\s+\S+\s+pause\s+\S+:0.3\n`))
			Consistently(session).ShouldNot(Say(`pause-`))
		})

		It("should print a specific revision in list format (absolute revision)", func() {
			workload.BumpImage(object)
			workload.BumpImage(object)

			session := RunPluginAndWait(append(args, "--revision=2")...)
			Eventually(session).Should(Say(`pause-\S+\s+2\s+\d/\d\s+\S+\n`))
			Consistently(session).ShouldNot(Say(`pause-`))
		})

		It("should print a specific revision in wide format (relative revision)", func() {
			workload.BumpImage(object)
			workload.BumpImage(object)

			session := RunPluginAndWait(append(args, "--revision=2", "-o", "wide")...)
			Eventually(session).Should(Say(`pause-\S+\s+2\s+\d/\d\s+\S+\s+pause\s+\S+:0.2\n`))
			Consistently(session).ShouldNot(Say(`pause-`))
		})

		It("should print a specific revision in yaml format", func() {
			workload.BumpImage(object)
			workload.BumpImage(object)

			session := RunPluginAndWait(append(args, "--revision=1", "-o", "yaml")...)

			yamlBytes, err := io.ReadAll(session.Out)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(yamlBytes)).To(ContainSubstring("image: " + workload.ImageRepository + ":0.1"))

			Expect(runtime.DecodeInto(decoder, yamlBytes, workload.RevisionObjectFor(object))).To(Succeed())
		})

		It("should print a specific revision's pod template in json format on --template-only", func() {
			workload.BumpImage(object)
			workload.BumpImage(object)

			session := RunPluginAndWait(append(args, "--revision=2", "-o", "json", "--template-only")...)

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

			session := RunPluginAndWait(append(args, "-o", "yaml")...)

			yamlBytes, err := io.ReadAll(session.Out)
			Expect(err).NotTo(HaveOccurred())

			list := &metav1.List{}
			Expect(runtime.DecodeInto(decoder, yamlBytes, list)).To(Succeed())
			Expect(list.Items).To(HaveLen(3))

			Expect(runtime.DecodeInto(decoder, list.Items[0].Raw, workload.RevisionObjectFor(object))).To(Succeed())
		})

		It("should list revisions of all resources in the namespace", func() {
			createObject(namespace, workload.AppName+"1")
			createObject(namespace, workload.AppName+"2")
			createObject(namespace, workload.AppName+"3")

			session := RunPluginAndWait(args[:len(args)-1]...)
			Eventually(session).Should(Say(`NAME\s+REVISION\s+READY\s+AGE\n`))
			Eventually(session).Should(Say(`pause-\S+\s+1\s+\d/\d\s+\S+\n`))
			Eventually(session).Should(Say(`pause1-\S+\s+1\s+\d/\d\s+\S+\n`))
			Eventually(session).Should(Say(`pause2-\S+\s+1\s+\d/\d\s+\S+\n`))
			Eventually(session).Should(Say(`pause3-\S+\s+1\s+\d/\d\s+\S+\n`))
		})

		It("should list revisions of label-selected resources in all namespaces", func() {
			createObject(namespace, workload.AppName+"1")

			namespace2 := workload.PrepareTestNamespace(namespace + "1")
			createObject(namespace2, workload.AppName+"2")
			createObject(namespace2, workload.AppName+"3")

			labelSelector := labels.SelectorFromValidatedSet(workload.CommonLabels())

			session := RunPluginAndWait("get", args[3], "-A", "-l", labelSelector.String())
			Eventually(session).Should(Say(`NAMESPACE\s+NAME\s+REVISION\s+READY\s+AGE\n`))
			Eventually(session).Should(Say(namespace + `\s+pause-\S+\s+1\s+\d/\d\s+\S+\n`))
			Eventually(session).Should(Say(namespace + `\s+pause1-\S+\s+1\s+\d/\d\s+\S+\n`))
			Eventually(session).Should(Say(namespace2 + `\s+pause2-\S+\s+1\s+\d/\d\s+\S+\n`))
			Eventually(session).Should(Say(namespace2 + `\s+pause3-\S+\s+1\s+\d/\d\s+\S+\n`))
		})
	}

	Context("Deployment", func() {
		BeforeEach(func() {
			object = workload.CreateDeployment(namespace, workload.AppName)
			args = append(args, "deployment", object.GetName())
		})

		testCommon(workload.CreateDeployment)

		It("should work with short type", func() {
			args[3] = "deploy"
			Eventually(RunPluginAndWait(args...)).Should(Say(`pause-\S+\s+1\s+\d/\d\s+\S+\n`))
		})

		It("should work with grouped type", func() {
			args[3] = "deployments.apps"
			Eventually(RunPluginAndWait(args...)).Should(Say(`pause-\S+\s+1\s+\d/\d\s+\S+\n`))
		})

		It("should work with fully-qualified type", func() {
			args[3] = "deployments.v1.apps"
			Eventually(RunPluginAndWait(args...)).Should(Say(`pause-\S+\s+1\s+\d/\d\s+\S+\n`))
		})

		It("should work with slash name", func() {
			args[3] = "deployment/pause"
			args = args[:len(args)-1]
			Eventually(RunPluginAndWait(args...)).Should(Say(`pause-\S+\s+1\s+\d/\d\s+\S+\n`))
		})

		It("should support the -v flag", func() {
			session := RunPluginAndWait(append(args, "-v6")...)
			Eventually(session.Err).Should(Say(`Config loaded from file`))
			Eventually(session.Err).Should(Say(`GET \S+/apis/apps/v1/namespaces/` + namespace + `/deployments/pause 200 OK`))
			Eventually(session.Err).Should(Say(`GET \S+/apis/apps/v1/namespaces/` + namespace + `/replicasets\?labelSelector=app%3Dpause%2Ce2e-test%3Dkubectl-revisions 200 OK`))
			Eventually(session).Should(Say(`NAME\s+REVISION\s+READY\s+AGE\n`))
			Eventually(session).Should(Say(`pause-\S+\s+1\s+\d/\d\s+\S+\n`))
		})

		It("should correctly print replicas", func() {
			workload.Scale(object, 2)
			Eventually(komega.Object(object)).Should(HaveField("Status.ReadyReplicas", int32(2)))

			// prepare second revision with broken image
			// this make it easy and deterministic to test the replica column for multiple revisions
			workload.SetImage(object, workload.ImageRepository+":non-existing")
			Eventually(komega.Object(object)).Should(HaveField("Status.UpdatedReplicas", int32(1)))

			session := RunPluginAndWait(args...)
			Eventually(session).Should(Say(`NAME\s+REVISION\s+READY\s+AGE\n`))
			Eventually(session).Should(Say(`pause-\S+\s+1\s+2/2\s+\S+\n`))
			Eventually(session).Should(Say(`pause-\S+\s+2\s+0/1\s+\S+\n`))
		})
	})

	Context("StatefulSet", func() {
		BeforeEach(func() {
			object = workload.CreateStatefulSet(namespace, workload.AppName)
			args = append(args, "statefulset", object.GetName())
		})

		testCommon(workload.CreateStatefulSet)

		It("should correctly print replicas", func() {
			workload.Scale(object, 2)
			Eventually(komega.Object(object)).Should(HaveField("Status.ReadyReplicas", int32(2)))

			// prepare second revision with broken image
			// this make it easy and deterministic to test the replica column for multiple revisions
			workload.SetImage(object, workload.ImageRepository+":non-existing")
			Eventually(komega.Object(object)).Should(HaveField("Status.UpdatedReplicas", int32(1)))

			session := RunPluginAndWait(args...)
			Eventually(session).Should(Say(`NAME\s+REVISION\s+READY\s+AGE\n`))
			Eventually(session).Should(Say(`pause-\S+\s+1\s+1/1\s+\S+\n`))
			Eventually(session).Should(Say(`pause-\S+\s+2\s+0/1\s+\S+\n`))
		})
	})

	Context("DaemonSet", func() {
		BeforeEach(func() {
			object = workload.CreateDaemonSet(namespace, workload.AppName)
			args = append(args, "daemonset", object.GetName())
		})

		testCommon(workload.CreateDaemonSet)

		It("should correctly print replicas", func() {
			Eventually(komega.Object(object)).Should(HaveField("Status.NumberReady", int32(3)))

			// prepare second revision with broken image
			// this make it easy and deterministic to test the replica column for multiple revisions
			workload.SetImage(object, workload.ImageRepository+":non-existing")
			Eventually(komega.Object(object)).Should(And(
				HaveField("Status.NumberReady", int32(2)),
				HaveField("Status.UpdatedNumberScheduled", int32(1)),
			))

			// We cannot determine from the DaemonSet status whether the old pod has finished terminating. To make testing
			// the plugin deterministic, wait until there are exactly 3 pods left in the namespace (stable state).
			// We don't need this for Deployment or StatefulSet as waiting for the status.updatedReplicas field ensures
			// the system has a stable state where nothing happens.
			Eventually(komega.ObjectList(&corev1.PodList{}, client.InNamespace(namespace))).Should(HaveField("Items", HaveLen(3)))

			session := RunPluginAndWait(args...)
			Eventually(session).Should(Say(`NAME\s+REVISION\s+READY\s+AGE\n`))
			Eventually(session).Should(Say(`pause-\S+\s+1\s+2/2\s+\S+\n`))
			Eventually(session).Should(Say(`pause-\S+\s+2\s+0/1\s+\S+\n`))
		})
	})
})
