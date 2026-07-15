package ui

import (
	"math"
	"testing"

	"gotest.tools/v3/assert"
)

const (
	MinInt32 = math.MinInt32
	MaxInt32 = math.MaxInt32
)

// Test clampedAdd[int32].
func TestClampedAddSigned(t *testing.T) {
	for _, tc := range []struct {
		x, y, minv, maxv, want int32
	}{
		{1, MaxInt32, MinInt32, MaxInt32, MaxInt32},
		{2, MaxInt32, MinInt32, MaxInt32, MaxInt32},
		{MaxInt32, 1, MinInt32, MaxInt32, MaxInt32},
		{MaxInt32, 2, MinInt32, MaxInt32, MaxInt32},
		{MaxInt32, MaxInt32, MinInt32, MaxInt32, MaxInt32},

		{-1, MinInt32, MinInt32, MaxInt32, MinInt32},
		{-2, MinInt32, MinInt32, MaxInt32, MinInt32},
		{MinInt32, -1, MinInt32, MaxInt32, MinInt32},
		{MinInt32, -2, MinInt32, MaxInt32, MinInt32},
		{MinInt32, MinInt32, MinInt32, MaxInt32, MinInt32},
	} {
		sign := "+"
		if tc.y < 0 {
			sign = ""
		}
		t.Logf("%d%s%d clamped at [%d,%d] == %d?",
			tc.x, sign, tc.y, tc.minv, tc.maxv, tc.want)

		assert.Equal(t, clampedAdd(tc.x, tc.y, tc.minv, tc.maxv), tc.want)
	}
}

// Test wrappedAdd[uint32].
func TestWrappedAddUnsigned(t *testing.T) {
	for _, tc := range []struct {
		x                uint32
		y                int32
		minv, maxv, want uint32
	}{
		//{0, 0, 0, 0, 0},
		//{0, 1, 0, 0, 0},
		//{0, -1, 0, 0, 0},
		{0, 1, 0, 1, 1},
		{0, 1, 0, 5, 1},
		{0, -1, 0, 5, 5},
		{0, -2, 0, 5, 4},
		{1, 2, 0, 5, 3},
		{5, 1, 0, 5, 0},
		{5, 2, 0, 5, 1},

		{0xffffffff, 1, 0, 0xffffffff, 0},
		{0xffffffff, 2, 0, 0xffffffff, 1},
		{0xffffffff, -1, 0, 0xffffffff, 0xfffffffe},
	} {
		sign := "+"
		if tc.y < 0 {
			sign = ""
		}
		t.Logf("%d%s%d wrapped at [%d,%d] == %d?",
			tc.x, sign, tc.y, tc.minv, tc.maxv, tc.want)

		assert.Equal(t, wrappedAdd(tc.x, tc.y, tc.minv, tc.maxv), tc.want)
	}
}
