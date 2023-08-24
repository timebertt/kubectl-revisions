package printer_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/timebertt/kubectl-history/pkg/history"
	. "github.com/timebertt/kubectl-history/pkg/printer"
)

var _ = Describe("RevisionPrinter", func() {
	var (
		p        *RevisionPrinter
		delegate *fakePrinter
	)

	BeforeEach(func() {
		delegate = &fakePrinter{}
		p = &RevisionPrinter{
			Delegate: delegate,
		}
	})

	It("should delegate unhandled objects", func() {
		obj := &corev1.ConfigMap{}
		Expect(p.PrintObj(obj, nil)).To(Succeed())

		Expect(delegate.printed).To(Equal(obj))
	})

	Context("printing a single revision", func() {
		var rev history.Revision

		BeforeEach(func() {
			var err error
			rev, err = history.NewReplicaSet(replicaSet(1))
			Expect(err).NotTo(HaveOccurred())
		})

		Context("TemplateOnly=false", func() {
			BeforeEach(func() {
				p.TemplateOnly = false
			})

			It("should print the full revision object", func() {
				Expect(p.PrintObj(rev, nil)).To(Succeed())

				Expect(delegate.printed).To(Equal(rev.Object()))
			})
		})

		Context("TemplateOnly=true", func() {
			BeforeEach(func() {
				p.TemplateOnly = true
			})

			It("should print only the pod template", func() {
				Expect(p.PrintObj(rev, nil)).To(Succeed())

				Expect(delegate.printed).To(Equal(rev.PodTemplate()))
			})
		})
	})

	Context("printing a list of revision", func() {
		var revs history.Revisions

		BeforeEach(func() {
			rev1, err := history.NewReplicaSet(replicaSet(1))
			Expect(err).NotTo(HaveOccurred())
			rev2, err := history.NewReplicaSet(replicaSet(2))
			Expect(err).NotTo(HaveOccurred())
			revs = history.Revisions{rev1, rev2}
		})

		Context("TemplateOnly=false", func() {
			BeforeEach(func() {
				p.TemplateOnly = false
			})

			It("should print the full revision objects", func() {
				Expect(p.PrintObj(revs, nil)).To(Succeed())

				Expect(delegate.printed).To(BeAssignableToTypeOf(&unstructured.UnstructuredList{}))

				list := delegate.printed.(*unstructured.UnstructuredList)
				Expect(list.Object).To(Equal(map[string]interface{}{
					"kind":       "List",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"resourceVersion": "",
					},
				}))

				Expect(list.Items).To(HaveExactElements([]unstructured.Unstructured{
					expectedUnstructuredReplicaSet(revs[0].Object().(*appsv1.ReplicaSet)),
					expectedUnstructuredReplicaSet(revs[1].Object().(*appsv1.ReplicaSet)),
				}))
			})
		})

		Context("TemplateOnly=true", func() {
			BeforeEach(func() {
				p.TemplateOnly = true
			})

			It("should print only the pod templates", func() {
				Expect(p.PrintObj(revs, nil)).To(Succeed())

				Expect(delegate.printed).To(BeAssignableToTypeOf(&unstructured.UnstructuredList{}))

				list := delegate.printed.(*unstructured.UnstructuredList)
				Expect(list.Object).To(Equal(map[string]interface{}{
					"kind":       "List",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"resourceVersion": "",
					},
				}))

				Expect(list.Items).To(HaveExactElements([]unstructured.Unstructured{
					expectedUnstructuredPod(revs[0].PodTemplate()),
					expectedUnstructuredPod(revs[1].PodTemplate()),
				}))
			})
		})
	})
})

func expectedUnstructuredReplicaSet(obj *appsv1.ReplicaSet) unstructured.Unstructured {
	obj.GetObjectKind().SetGroupVersionKind(appsv1.SchemeGroupVersion.WithKind("ReplicaSet"))
	unstructuredContent, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	Expect(err).NotTo(HaveOccurred())
	return unstructured.Unstructured{Object: unstructuredContent}
}

func expectedUnstructuredPod(obj *corev1.Pod) unstructured.Unstructured {
	unstructuredContent, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	Expect(err).NotTo(HaveOccurred())
	return unstructured.Unstructured{Object: unstructuredContent}
}
