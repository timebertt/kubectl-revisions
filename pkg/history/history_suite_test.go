package history_test

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gcustom"
	"github.com/onsi/gomega/types"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	. "github.com/timebertt/kubectl-revisions/pkg/history"
)

func TestHistory(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "History Suite")
}

func haveNumber(expected int64) types.GomegaMatcher {
	return gcustom.MakeMatcher(func(rev Revision) (bool, error) {
		return rev.Number() == expected, nil
	}).WithMessage(fmt.Sprintf("have revision %d", expected))
}

func copyMap[K comparable, V any](in map[K]V) map[K]V {
	if in == nil {
		return nil
	}

	out := make(map[K]V, len(in))
	for k, v := range in {
		out[k] = v
	}

	return out
}

func beNotFoundError() types.GomegaMatcher {
	return gcustom.MakeMatcher(func(err error) (bool, error) {
		return apierrors.IsNotFound(err), nil
	}).WithMessage("be NotFound error")
}
