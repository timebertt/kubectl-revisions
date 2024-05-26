package workload

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var testClient client.Client

func SetClient(c client.Client) {
	testClient = c
}
