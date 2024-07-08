package history

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

// Scheme is a *runtime.Scheme including all types necessary to handle history-related objects.
var Scheme = runtime.NewScheme()
var localSchemeBuilder = runtime.SchemeBuilder{
	corev1.AddToScheme,
	appsv1.AddToScheme,
}

// DecodingVersions is a list of GroupVersions that need to be decoded for handling history-related objects.
var DecodingVersions = []schema.GroupVersion{
	corev1.SchemeGroupVersion,
	appsv1.SchemeGroupVersion,
}

// AddToScheme adds all types necessary to handle history-related objects to the given scheme.
var AddToScheme = localSchemeBuilder.AddToScheme

func init() {
	metav1.AddToGroupVersion(Scheme, schema.GroupVersion{Version: "v1"})
	utilruntime.Must(AddToScheme(Scheme))
}
