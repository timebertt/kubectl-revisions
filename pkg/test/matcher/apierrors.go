package matcher

import (
	"fmt"

	"github.com/onsi/gomega/gcustom"
	"github.com/onsi/gomega/types"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func BeNotFoundError() types.GomegaMatcher {
	return gcustom.MakeMatcher(func(err error) (bool, error) {
		return apierrors.IsNotFound(err), nil
	}).WithMessage(fmt.Sprintf("be %q error", metav1.StatusReasonNotFound))
}
