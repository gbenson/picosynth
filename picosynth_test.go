package picosynth

import (
	"fmt"
	"math"

	"gotest.tools/v3/assert/cmp"
)

const (
	DefaultAbsoluteTolerance = 1e-12
	DefaultRelativeTolerance = 1e-6
)

// NearlyEqual returns a [cmp.Comparison] that succeeds if x ≈ y.
func NearlyEqual(x, y float64) cmp.Comparison {
	return func() cmp.Result {
		if x == y {
			return cmp.ResultSuccess
		}

		delta := math.Abs(y - x)
		if delta < DefaultAbsoluteTolerance {
			return cmp.ResultSuccess
		}

		if x != 0 && delta/math.Abs(x) < DefaultRelativeTolerance {
			return cmp.ResultSuccess
		}

		return cmp.ResultFailure(
			fmt.Sprintf("%v !≈ %v (delta = %v)", x, y, delta),
		)
	}
}
