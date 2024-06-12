package e2e

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/timebertt/kubectl-revisions/test/e2e/exec"
	"github.com/timebertt/kubectl-revisions/test/e2e/workload"
)

var _ = Describe("diff command", func() {
	var (
		namespace string
		object    client.Object

		args []string
	)

	BeforeEach(func() {
		namespace = workload.PrepareTestNamespace()
		args = []string{"diff", "-n", namespace}
	})

	Describe("command aliases", func() {
		BeforeEach(func() {
			object = workload.CreateDeployment(namespace)
			args = append(args, "deployment", object.GetName())
		})

		It("should work with alias why", func() {
			args[0] = "why"

			workload.BumpImage(object)

			session := RunPluginAndWait(args...)
			Eventually(session).Should(Say(`--- \S+\/1-pause-\S+\s`))
			Eventually(session).Should(Say(`\+\+\+ \S+\/2-pause-\S+\s`))
			Eventually(session).Should(Say(`-.+:0.1\n`))
			Eventually(session).Should(Say(`\+.+:0.2\n`))
		})
	})

	testCommon := func() {
		It("should diff the last two revisions", func() {
			workload.BumpImage(object)
			workload.BumpImage(object)

			session := RunPluginAndWait(args...)
			Eventually(session).Should(Say(`--- \S+\/2-pause-\S+\s`))
			Eventually(session).Should(Say(`\+\+\+ \S+\/3-pause-\S+\s`))
			Eventually(session).Should(Say(`-.+:0.2\n`))
			Eventually(session).Should(Say(`\+.+:0.3\n`))
		})

		It("should diff the given revisions", func() {
			workload.BumpImage(object)
			workload.BumpImage(object)

			session := RunPluginAndWait(append(args, "--revision=1,3")...)
			Eventually(session).Should(Say(`--- \S+\/1-pause-\S+\s`))
			Eventually(session).Should(Say(`\+\+\+ \S+\/3-pause-\S+\s`))
			Eventually(session).Should(Say(`-.+:0.1\n`))
			Eventually(session).Should(Say(`\+.+:0.3\n`))
		})

		It("should diff the given revision and its predecessor (absolute)", func() {
			workload.BumpImage(object)
			workload.BumpImage(object)

			session := RunPluginAndWait(append(args, "--revision=2")...)
			Eventually(session).Should(Say(`--- \S+\/1-pause-\S+\s`))
			Eventually(session).Should(Say(`\+\+\+ \S+\/2-pause-\S+\s`))
			Eventually(session).Should(Say(`-.+:0.1\n`))
			Eventually(session).Should(Say(`\+.+:0.2\n`))
		})

		It("should diff the given revision and its predecessor (relative)", func() {
			workload.BumpImage(object)
			workload.BumpImage(object)

			session := RunPluginAndWait(append(args, "--revision=-2")...)
			Eventually(session).Should(Say(`--- \S+\/1-pause-\S+\s`))
			Eventually(session).Should(Say(`\+\+\+ \S+\/2-pause-\S+\s`))
			Eventually(session).Should(Say(`-.+:0.1\n`))
			Eventually(session).Should(Say(`\+.+:0.2\n`))
		})

		It("should diff the revisions in the given format", func() {
			workload.BumpImage(object)
			workload.BumpImage(object)

			session := RunPluginAndWait(append(args, "-o", "jsonpath={.spec.containers[0].image}")...)
			Eventually(session).Should(Say(`--- \S+\/2-pause-\S+\s`))
			Eventually(session).Should(Say(`\+\+\+ \S+\/3-pause-\S+\s`))
			Eventually(session).Should(Say(`-.+:0.2\n`))
			Eventually(session).Should(Say(`\+.+:0.3\n`))
		})

		Context("external diff", func() {
			It("should invoke the external diff program", func() {
				workload.BumpImage(object)

				cmd := NewPluginCommand(args...)
				cmd.Env = append(cmd.Env, "KUBECTL_EXTERNAL_DIFF=ls")

				session := Wait(RunCommand(cmd))
				Eventually(session).Should(Say(`/1-pause-`))
				Eventually(session).Should(Say(`.` + namespace + `.pause\n`))
				Eventually(session).Should(Say(`/2-pause-`))
				Eventually(session).Should(Say(`.` + namespace + `.pause\n`))
			})

			It("should invoke dyff as external diff program with subcommand and flags", func() {
				workload.BumpImage(object)

				cmd := NewPluginCommand(args...)
				// dyff has some special integration in place that makes it compatible with `kubectl diff`. See
				// https://github.com/homeport/dyff/pull/149
				// Add a dedicated test that ensures `kubectl revisions diff` works with the recommended setup for using `dyff`
				// with kubectl, i.e., that it works with the same KUBECTL_EXTERNAL_DIFF setting.
				cmd.Env = append(cmd.Env, "KUBECTL_EXTERNAL_DIFF=dyff between --omit-header --set-exit-code")

				session := Wait(RunCommand(cmd))
				Eventually(session).Should(Say(`spec.containers.pause.image\n`))
				Eventually(session).Should(Say(`value change\n`))
				Eventually(session).Should(Say(`-.+:0.1\n`))
				Eventually(session).Should(Say(`\+.+:0.2\n`))
			})
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

			workload.BumpImage(object)

			session := RunPluginAndWait(args...)
			Eventually(session).Should(Say(`--- \S+\/1-pause-\S+\s`))
			Eventually(session).Should(Say(`\+\+\+ \S+\/2-pause-\S+\s`))
			Eventually(session).Should(Say(`-.+:0.1\n`))
			Eventually(session).Should(Say(`\+.+:0.2\n`))
		})

		It("should work with grouped type", func() {
			args[3] = "deployments.apps"

			workload.BumpImage(object)

			session := RunPluginAndWait(args...)
			Eventually(session).Should(Say(`--- \S+\/1-pause-\S+\s`))
			Eventually(session).Should(Say(`\+\+\+ \S+\/2-pause-\S+\s`))
			Eventually(session).Should(Say(`-.+:0.1\n`))
			Eventually(session).Should(Say(`\+.+:0.2\n`))
		})

		It("should work with fully-qualified type", func() {
			args[3] = "deployments.v1.apps"

			workload.BumpImage(object)

			session := RunPluginAndWait(args...)
			Eventually(session).Should(Say(`--- \S+\/1-pause-\S+\s`))
			Eventually(session).Should(Say(`\+\+\+ \S+\/2-pause-\S+\s`))
			Eventually(session).Should(Say(`-.+:0.1\n`))
			Eventually(session).Should(Say(`\+.+:0.2\n`))
		})

		It("should work with slash name", func() {
			args[3] = "deployment/pause"
			args = args[:len(args)-1]

			workload.BumpImage(object)

			session := RunPluginAndWait(args...)
			Eventually(session).Should(Say(`--- \S+\/1-pause-\S+\s`))
			Eventually(session).Should(Say(`\+\+\+ \S+\/2-pause-\S+\s`))
			Eventually(session).Should(Say(`-.+:0.1\n`))
			Eventually(session).Should(Say(`\+.+:0.2\n`))
		})

		It("should diff the full revision objects on --template-only=false", func() {
			workload.BumpImage(object)

			session := Wait(RunCommand(NewPluginCommand(append(args, "--template-only=false")...)))
			Eventually(session).Should(Say(`-.+deployment.kubernetes.io/revision: "1"`))
			Eventually(session).Should(Say(`\+.+deployment.kubernetes.io/revision: "2"`))
			Eventually(session).Should(Say(`-.+pod-template-hash: `))
			Eventually(session).Should(Say(`\+.+pod-template-hash: `))
			Eventually(session).Should(Say(`-.+:0.1\n`))
			Eventually(session).Should(Say(`\+.+:0.2\n`))
		})
	})

	Context("StatefulSet", func() {
		BeforeEach(func() {
			object = workload.CreateStatefulSet(namespace)
			args = append(args, "statefulset", object.GetName())
		})

		testCommon()

		It("should diff the full revision objects on --template-only=false", func() {
			workload.BumpImage(object)

			session := Wait(RunCommand(NewPluginCommand(append(args, "--template-only=false")...)))
			Eventually(session).Should(Say(`-.+:0.1\n`))
			Eventually(session).Should(Say(`\+.+:0.2\n`))
			Eventually(session).Should(Say(`-.+controller.kubernetes.io/hash: `))
			Eventually(session).Should(Say(`\+.+controller.kubernetes.io/hash: `))
			Eventually(session).Should(Say(`-revision: 1`))
			Eventually(session).Should(Say(`\+revision: 2`))
		})
	})

	Context("DaemonSet", func() {
		BeforeEach(func() {
			object = workload.CreateDaemonSet(namespace)
			args = append(args, "daemonset", object.GetName())
		})

		testCommon()

		It("should diff the full revision objects on --template-only=false", func() {
			workload.BumpImage(object)

			session := Wait(RunCommand(NewPluginCommand(append(args, "--template-only=false")...)))
			Eventually(session).Should(Say(`-.+:0.1\n`))
			Eventually(session).Should(Say(`\+.+:0.2\n`))
			Eventually(session).Should(Say(`-.+controller-revision-hash: `))
			Eventually(session).Should(Say(`\+.+controller-revision-hash: `))
			Eventually(session).Should(Say(`-revision: 1`))
			Eventually(session).Should(Say(`\+revision: 2`))
		})
	})
})
