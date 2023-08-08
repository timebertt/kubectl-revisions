package diff_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDiff(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Diff Suite")
}
