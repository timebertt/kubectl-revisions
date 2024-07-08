package history_test

import (
	"context"
	"fmt"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	deploymentutil "k8s.io/kubectl/pkg/util/deployment"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	. "github.com/timebertt/kubectl-revisions/pkg/history"
)

var _ = Describe("DeploymentHistory", func() {
	var (
		ctx        context.Context
		fakeClient client.Client
	)

	BeforeEach(func() {
		ctx = context.Background()
		fakeClient = fakeclient.NewClientBuilder().Build()
	})

	Describe("initialization", func() {
		It("should be constructable via For", func() {
			history, err := For(fakeClient, &appsv1.Deployment{})
			Expect(err).NotTo(HaveOccurred())
			Expect(history).NotTo(BeNil())
			Expect(history).To(BeAssignableToTypeOf(DeploymentHistory{}))

			h := history.(DeploymentHistory)
			Expect(h.Client).To(Equal(fakeClient))
		})

		It("should be constructable via ForGroupKind", func() {
			history, err := ForGroupKind(fakeClient, appsv1.SchemeGroupVersion.WithKind("Deployment").GroupKind())
			Expect(err).NotTo(HaveOccurred())
			Expect(history).NotTo(BeNil())
			Expect(history).To(BeAssignableToTypeOf(DeploymentHistory{}))

			h := history.(DeploymentHistory)
			Expect(h.Client).To(Equal(fakeClient))
		})
	})

	Describe("ListRevisions", func() {
		var (
			history DeploymentHistory

			deployment *appsv1.Deployment

			replicaSet1, replicaSet3, replicaSetUnrelated *appsv1.ReplicaSet
		)

		BeforeEach(func() {
			history = DeploymentHistory{
				Client: fakeClient,
			}

			deployment = &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "deploy",
					Namespace: "test",
					Labels: map[string]string{
						"app": "deploy",
					},
				},
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "deploy",
						},
					},
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{
								Name: "test",
							}},
						},
					},
				},
			}

			Expect(fakeClient.Create(ctx, deployment)).To(Succeed())

			replicaSet1 = replicaSetForDeployment(deployment, 1, fakeClient.Scheme())
			replicaSet1.Name = "app-b"
			Expect(fakeClient.Create(ctx, replicaSet1)).To(Succeed())
			replicaSet1.Status.Replicas = 1
			replicaSet1.Status.ReadyReplicas = 1
			Expect(fakeClient.Status().Update(ctx, replicaSet1)).To(Succeed())

			// create a non-sorted list of ReplicaSets to verify that ListRevisions returns a sorted list
			replicaSet3 = replicaSetForDeployment(deployment, 3, fakeClient.Scheme())
			replicaSet3.Name = "app-a"
			Expect(fakeClient.Create(ctx, replicaSet3)).To(Succeed())
			replicaSet3.Status.Replicas = 1
			replicaSet3.Status.ReadyReplicas = 0
			Expect(fakeClient.Status().Update(ctx, replicaSet3)).To(Succeed())

			replicaSetUnrelated = replicaSetForDeployment(deployment, 0, fakeClient.Scheme())
			replicaSetUnrelated.OwnerReferences[0].UID = "other"
			Expect(fakeClient.Create(ctx, replicaSetUnrelated)).To(Succeed())
		})

		It("should return an empty list if there are no ReplicaSets", func() {
			deployment.ResourceVersion = ""
			deployment.UID = ""
			deployment.Namespace = "other"
			Expect(fakeClient.Create(ctx, deployment)).To(Succeed())

			revs, err := history.ListRevisions(ctx, deployment)
			Expect(err).NotTo(HaveOccurred())
			Expect(revs).To(BeEmpty())
		})

		It("should return a sorted list of the owned ReplicaSets", func() {
			revs, err := history.ListRevisions(ctx, deployment)
			Expect(err).NotTo(HaveOccurred())
			Expect(revs).To(HaveLen(2))

			Expect(revs[0].Number()).To(BeEquivalentTo(1))
			Expect(revs[0].Object()).To(Equal(replicaSet1))
			Expect(revs[0].CurrentReplicas()).To(BeEquivalentTo(1))
			Expect(revs[0].ReadyReplicas()).To(BeEquivalentTo(1))

			Expect(revs[1].Number()).To(BeEquivalentTo(3))
			Expect(revs[1].Object()).To(Equal(replicaSet3))
			Expect(revs[1].CurrentReplicas()).To(BeEquivalentTo(1))
			Expect(revs[1].ReadyReplicas()).To(BeEquivalentTo(0))
		})

		It("should also work via ListRevisions shortcut", func() {
			revs, err := ListRevisions(ctx, fakeClient, deployment)
			Expect(err).NotTo(HaveOccurred())
			Expect(revs).To(HaveLen(2))

			Expect(revs[0].Object()).To(Equal(replicaSet1))
			Expect(revs[1].Object()).To(Equal(replicaSet3))
		})
	})
})

func replicaSetForDeployment(deployment *appsv1.Deployment, revision int64, scheme *runtime.Scheme) *appsv1.ReplicaSet {
	labels := copyMap(deployment.Spec.Selector.MatchLabels)
	labels[appsv1.DefaultDeploymentUniqueLabelKey] = strconv.FormatInt(revision, 10)

	template := deployment.Spec.Template.DeepCopy()
	template.Labels = labels
	template.Spec.Containers[0].Image = fmt.Sprintf("test:%d", revision)

	replicaSet := &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%d", deployment.Name, revision),
			Namespace: deployment.Namespace,
			Labels:    labels,
			Annotations: map[string]string{
				deploymentutil.RevisionAnnotation: strconv.FormatInt(revision, 10),
			},
		},
		Spec: appsv1.ReplicaSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
		},
	}

	Expect(controllerutil.SetControllerReference(deployment, replicaSet, scheme)).To(Succeed())

	return replicaSet
}
