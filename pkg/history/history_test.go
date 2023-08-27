package history_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	. "github.com/timebertt/kubectl-history/pkg/history"
	"github.com/timebertt/kubectl-history/pkg/history/fake"
)

var _ = Describe("History", func() {
	var fakeClient client.Client

	BeforeEach(func() {
		fakeClient = fakeclient.NewClientBuilder().Build()
	})

	Describe("For", func() {
		It("should fail if the object's GVK can't be determined", func() {
			scheme := runtime.NewScheme()
			Expect(appsv1.AddToScheme(scheme)).To(Succeed())
			c := fakeclient.NewClientBuilder().WithScheme(scheme).Build()

			history, err := For(c, &corev1.ConfigMap{})
			Expect(history).To(BeNil())
			Expect(runtime.IsNotRegisteredError(err)).To(BeTrue())
		})

		It("should fail if the object's GVK is not supported", func() {
			history, err := For(fakeClient, &corev1.ConfigMap{})
			Expect(history).To(BeNil())
			Expect(err).To(MatchError("ConfigMap is not supported"))
		})
	})

	Describe("ForGroupKind", func() {
		It("should fail if the GVK is not supported", func() {
			history, err := ForGroupKind(fakeClient, corev1.SchemeGroupVersion.WithKind("ConfigMap").GroupKind())
			Expect(history).To(BeNil())
			Expect(err).To(MatchError("ConfigMap is not supported"))
		})
	})
})

var _ = Describe("Revisions", func() {
	var (
		revs Revisions
	)

	BeforeEach(func() {
		revs = Revisions{someRevision(1), someRevision(2), someRevision(4)}
	})

	Describe("GetObjectKind", func() {
		It("should return an empty ObjectKind if the list is empty", func() {
			revs = nil
			Expect(revs.GetObjectKind().GroupVersionKind().Empty()).To(BeTrue())
		})

		It("should return the first revision's ObjectKind", func() {
			revs[1].GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{Kind: "Foo"})
			Expect(revs.GetObjectKind().GroupVersionKind()).To(Equal(schema.GroupVersionKind{
				Version: "v1",
				Kind:    "Pod",
			}))
		})
	})

	Describe("DeepCopyObject", func() {
		It("should return nil if the list is nil", func() {
			revs = nil
			Expect(revs.DeepCopyObject()).To(BeNil())
		})

		It("should return a copy of the list", func() {
			copied := revs.DeepCopyObject()
			Expect(copied).To(BeAssignableToTypeOf(revs))
			Expect(copied).To(HaveExactElements(revs[0], revs[1], revs[2]))

			copiedRevs := copied.(Revisions)
			Expect(copiedRevs[0]).NotTo(BeIdenticalTo(revs[0]))
			Expect(copiedRevs[1]).NotTo(BeIdenticalTo(revs[1]))
			Expect(copiedRevs[2]).NotTo(BeIdenticalTo(revs[2]))
		})
	})

	Describe("ByNumber", func() {
		It("should return an error if the list is empty", func() {
			revs = nil
			revision, err := revs.ByNumber(1)
			Expect(revision).To(BeNil())
			Expect(err).To(MatchError("revision 1 not found"))
		})

		It("should return an error if revision 0 is requested", func() {
			revision, err := revs.ByNumber(0)
			Expect(revision).To(BeNil())
			Expect(err).To(MatchError("invalid revision number 0"))
		})

		Context("positive number", func() {
			It("should return an error if the revision is not found", func() {
				revision, err := revs.ByNumber(3)
				Expect(revision).To(BeNil())
				Expect(err).To(MatchError("revision 3 not found"))
			})

			It("should return the correct revision", func() {
				Expect(revs.ByNumber(1)).To(haveNumber(1))
				Expect(revs.ByNumber(2)).To(haveNumber(2))
				Expect(revs.ByNumber(4)).To(haveNumber(4))
			})
		})

		Context("negative number", func() {
			It("should return an error if the number's absolute is smaller than the list's length", func() {
				revision, err := revs.ByNumber(-4)
				Expect(revision).To(BeNil())
				Expect(err).To(MatchError("revision -4 not found"))
			})

			It("should return the correct revision", func() {
				Expect(revs.ByNumber(-3)).To(haveNumber(1))
				Expect(revs.ByNumber(-2)).To(haveNumber(2))
				Expect(revs.ByNumber(-1)).To(haveNumber(4))
			})
		})
	})

	Describe("Predecessor", func() {
		It("should return an error if the list is empty", func() {
			revs = nil
			revision, err := revs.Predecessor(1)
			Expect(revision).To(BeNil())
			Expect(err).To(MatchError("revision 1 not found"))
		})

		It("should return an error if revision 0 is requested", func() {
			revision, err := revs.Predecessor(0)
			Expect(revision).To(BeNil())
			Expect(err).To(MatchError("invalid revision number 0"))
		})

		Context("positive number", func() {
			It("should return an error if the revision is not found", func() {
				revision, err := revs.Predecessor(3)
				Expect(revision).To(BeNil())
				Expect(err).To(MatchError("revision 3 not found"))
			})

			It("should return the correct revision", func() {
				Expect(revs.Predecessor(2)).To(haveNumber(1))
				Expect(revs.Predecessor(4)).To(haveNumber(2))
			})

			It("should return an error if the revision doesn't have a predecessor", func() {
				revision, err := revs.Predecessor(1)
				Expect(revision).To(BeNil())
				Expect(err).To(MatchError("predecessor of revision 1 not found"))
			})
		})

		Context("negative number", func() {
			It("should return an error if the number's absolute is smaller than the list's length", func() {
				revision, err := revs.Predecessor(-4)
				Expect(revision).To(BeNil())
				Expect(err).To(MatchError("revision -4 not found"))
			})

			It("should return the correct revision", func() {
				Expect(revs.Predecessor(-2)).To(haveNumber(1))
				Expect(revs.Predecessor(-1)).To(haveNumber(2))
			})

			It("should return an error if the revision doesn't have a predecessor", func() {
				revision, err := revs.Predecessor(-3)
				Expect(revision).To(BeNil())
				Expect(err).To(MatchError("predecessor of revision 1 not found"))
			})
		})
	})
})

func someRevision(num int64) Revision {
	return &fake.Revision{
		Num: num,
		Obj: &corev1.Pod{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Pod",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("test-%d", num),
			},
		},
	}
}
