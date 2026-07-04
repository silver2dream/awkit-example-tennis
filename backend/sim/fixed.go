package sim

import (
	"math/bits"

	"tennis-arena-backend/sim/constants"
)

type Fixed int64

const (
	FixedShift = constants.FixedShift
	FixedOne   = Fixed(constants.FixedOne)
	FixedZero  = Fixed(0)
)

func FixedFromInt(v int64) Fixed {
	return Fixed(v << FixedShift)
}

func (f Fixed) Mul(other Fixed) Fixed {
	if f == 0 || other == 0 {
		return 0
	}

	negative := (f < 0) != (other < 0)
	hi, lo := bits.Mul64(absFixedRaw(f), absFixedRaw(other))

	const overflowHigh = uint64(1) << (64 - FixedShift)
	if hi >= overflowHigh {
		if negative {
			return Fixed(-1 << 63)
		}
		return Fixed(1<<63 - 1)
	}

	raw := (hi << FixedShift) | (lo >> FixedShift)

	const maxInt64 = uint64(1<<63 - 1)
	const minInt64Abs = uint64(1) << 63

	if negative {
		if raw >= minInt64Abs {
			return Fixed(-1 << 63)
		}
		return Fixed(-int64(raw))
	}
	if raw > maxInt64 {
		return Fixed(1<<63 - 1)
	}
	return Fixed(int64(raw))
}

func ClampFixed(v, min, max Fixed) Fixed {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func AbsFixed(v Fixed) Fixed {
	if v == Fixed(-1<<63) {
		return Fixed(1<<63 - 1)
	}
	if v < 0 {
		return -v
	}
	return v
}

func clampUnit(v Fixed) Fixed {
	return ClampFixed(v, 0, FixedOne)
}

func clampSignedUnit(v Fixed) Fixed {
	return ClampFixed(v, -FixedOne, FixedOne)
}

func absFixedRaw(v Fixed) uint64 {
	if v >= 0 {
		return uint64(v)
	}
	return uint64(^int64(v)) + 1
}
