package files

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_onSegment(t *testing.T) {
	type Test struct {
		a1  uint32
		a2  uint32
		b1  uint32
		b2  uint32
		exp bool
	}
	testFn := func(t *testing.T, tc Test) {
		assert.Equal(t, tc.exp, onSegment(tc.a1, tc.a2, tc.b1, tc.b2))
	}

	testCases := []Test{
		{
			a1:  1,
			a2:  3,
			b1:  2,
			b2:  4,
			exp: true,
		},
		{
			a1:  4,
			a2:  7,
			b1:  4,
			b2:  7,
			exp: true,
		},
		{
			a1:  3,
			a2:  5,
			b1:  2,
			b2:  4,
			exp: true,
		},
		{
			a1:  2,
			a2:  4,
			b1:  3,
			b2:  5,
			exp: true,
		},
		{
			a1:  1,
			a2:  3,
			b1:  4,
			b2:  5,
			exp: false,
		},
		{
			a1:  8,
			a2:  9,
			b1:  6,
			b2:  7,
			exp: false,
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			testFn(t, tc)
		})
	}

}
