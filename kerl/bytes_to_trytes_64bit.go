// +build amd64 arm64 mips64

// Package kerl implements the Kerl hashing function.
package kerl

import (
	"unsafe"

	. "github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/kerl/bigint"
)

func bytesToTryteValues(bytes []byte) []int8 {
	// copy and convert bytes to bigint
	rb := make([]byte, len(bytes))
	copy(rb, bytes)
	bigint.Reverse(rb)
	b := (*(*[]uint32)(unsafe.Pointer(&rb)))[0:IntLength]

	// the two's complement representation is only correct, if the number fits
	// into 48 bytes, i.e. has the 243th trit set to 0
	bigintZeroLastTrit(b)

	// convert to the unsigned bigint representing non-balanced ternary
	bigint.MustAdd(b, halfThree)

	vs := make([]int8, HashTrytesSize)

	// initially, all words of the bigint are non-zero
	nzIndex := IntLength - 1
	for i := 0; i < HashTrytesSize-1; i++ {
		// divide the bigint by the radix
		var rem uint32
		for i := nzIndex; i >= 0; i-- {
			v := (uint64(rem) << 32) | uint64(b[i])
			b[i], rem = uint32(v/tryteRadix), uint32(v%tryteRadix)
		}
		// the tryte value is the remainder converted back to balanced ternary
		vs[i] = int8(rem) - halfTryte

		// decrement index, if the highest considered word of the bigint turned zero
		if nzIndex > 0 && b[nzIndex] == 0 {
			nzIndex--
		}
	}

	// special case for the last tryte, where no further division is necessary
	vs[HashTrytesSize-1] = tryteZeroLastTrit(int8(b[0]) - halfTryte)

	return vs
}
