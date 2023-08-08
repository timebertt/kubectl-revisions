package runutil_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRunutil(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Runutil Suite")
}
