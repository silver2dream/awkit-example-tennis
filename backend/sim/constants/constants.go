package constants

const (
	FixedShift = 32
	FixedOne   = int64(1) << FixedShift
)

const (
	StrokeFlatBaseSpeed    = 38 * FixedOne
	StrokeTopspinBaseSpeed = 34 * FixedOne
	StrokeSliceBaseSpeed   = 26 * FixedOne
)

const (
	StrokeFlatTopSpin    = 2 * FixedOne
	StrokeTopspinTopSpin = 72 * FixedOne
	StrokeSliceBackSpin  = -54 * FixedOne
)

const (
	StrokeSideSpinMax       = 18 * FixedOne
	StrokeAimLateralScale   = FixedOne / 4
	StrokeTimingPowerLoss   = FixedOne / 2
	StrokeTimingAimErrorMax = FixedOne / 4
)
