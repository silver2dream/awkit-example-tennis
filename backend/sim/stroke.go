package sim

import "tennis-arena-backend/sim/constants"

type SwingType uint8

const (
	SwingFlat SwingType = iota
	SwingTopspin
	SwingSlice
)

type StrokeInput struct {
	Swing         SwingType
	Aim           Fixed
	Power         Fixed
	TimingQuality Fixed
}

type StrokeOutput struct {
	Velocity        Vector3
	AngularVelocity Vector3
}

// ApplyStroke uses X as lateral, Y as vertical, and Z as forward. Angular X is
// top/back spin, and angular Y is sidespin.
func ApplyStroke(input StrokeInput) StrokeOutput {
	power := clampUnit(input.Power)
	if power == 0 {
		return StrokeOutput{}
	}

	quality := clampUnit(input.TimingQuality)
	miss := FixedOne - quality
	effectivePower := power.Mul(timingPowerFactor(miss))
	if effectivePower == 0 {
		return StrokeOutput{}
	}

	aim := clampSignedUnit(input.Aim + deterministicAimError(miss))
	baseSpeed, topSpin := strokeProfile(input.Swing)
	forward := baseSpeed.Mul(effectivePower)
	lateral := forward.Mul(aim).Mul(Fixed(constants.StrokeAimLateralScale))

	return StrokeOutput{
		Velocity: Vector3{
			X: lateral,
			Y: 0,
			Z: forward,
		},
		AngularVelocity: Vector3{
			X: topSpin.Mul(effectivePower),
			Y: Fixed(constants.StrokeSideSpinMax).Mul(aim).Mul(effectivePower),
			Z: 0,
		},
	}
}

func Stroke(input StrokeInput) StrokeOutput {
	return ApplyStroke(input)
}

func strokeProfile(swing SwingType) (speed Fixed, topSpin Fixed) {
	switch swing {
	case SwingTopspin:
		return Fixed(constants.StrokeTopspinBaseSpeed), Fixed(constants.StrokeTopspinTopSpin)
	case SwingSlice:
		return Fixed(constants.StrokeSliceBaseSpeed), Fixed(constants.StrokeSliceBackSpin)
	case SwingFlat:
		fallthrough
	default:
		return Fixed(constants.StrokeFlatBaseSpeed), Fixed(constants.StrokeFlatTopSpin)
	}
}

func timingPowerFactor(miss Fixed) Fixed {
	return FixedOne - miss.Mul(Fixed(constants.StrokeTimingPowerLoss))
}

func deterministicAimError(miss Fixed) Fixed {
	return miss.Mul(Fixed(constants.StrokeTimingAimErrorMax))
}
