package sim

import (
	"testing"

	"tennis-arena-backend/sim/constants"
)

func TestApplyStrokeSwingProfiles(t *testing.T) {
	input := StrokeInput{
		Power:         FixedOne,
		TimingQuality: FixedOne,
	}

	flat := ApplyStroke(withSwing(input, SwingFlat))
	topspin := ApplyStroke(withSwing(input, SwingTopspin))
	slice := ApplyStroke(withSwing(input, SwingSlice))

	if flat.Velocity.Z <= topspin.Velocity.Z {
		t.Fatalf("flat forward speed = %d, want greater than topspin %d", flat.Velocity.Z, topspin.Velocity.Z)
	}
	if flat.Velocity.Z != Fixed(constants.StrokeFlatBaseSpeed) {
		t.Fatalf("perfect timing flat speed = %d, want base speed %d", flat.Velocity.Z, constants.StrokeFlatBaseSpeed)
	}
	if flat.Velocity.X != 0 {
		t.Fatalf("perfect center aim lateral velocity = %d, want on-target zero", flat.Velocity.X)
	}
	if topspin.Velocity.Z <= slice.Velocity.Z {
		t.Fatalf("topspin forward speed = %d, want greater than slice %d", topspin.Velocity.Z, slice.Velocity.Z)
	}
	if AbsFixed(flat.AngularVelocity.X) >= AbsFixed(topspin.AngularVelocity.X) {
		t.Fatalf("flat top spin = %d, want lower than topspin %d", flat.AngularVelocity.X, topspin.AngularVelocity.X)
	}
	if topspin.AngularVelocity.X <= 0 {
		t.Fatalf("topspin angular X = %d, want positive topspin", topspin.AngularVelocity.X)
	}
	if slice.AngularVelocity.X >= 0 {
		t.Fatalf("slice angular X = %d, want negative backspin", slice.AngularVelocity.X)
	}
	if AbsFixed(slice.AngularVelocity.X) <= AbsFixed(flat.AngularVelocity.X) {
		t.Fatalf("slice backspin = %d, want greater magnitude than flat spin %d", slice.AngularVelocity.X, flat.AngularVelocity.X)
	}
	if flat.AngularVelocity.Y != 0 || topspin.AngularVelocity.Y != 0 || slice.AngularVelocity.Y != 0 {
		t.Fatalf("center aim should not add sidespin: flat=%d topspin=%d slice=%d", flat.AngularVelocity.Y, topspin.AngularVelocity.Y, slice.AngularVelocity.Y)
	}
}

func TestApplyStrokeAimControlsLateralDirectionAndSideSpin(t *testing.T) {
	base := StrokeInput{
		Swing:         SwingTopspin,
		Power:         FixedOne,
		TimingQuality: FixedOne,
	}

	right := ApplyStroke(withAim(base, FixedOne/2))
	left := ApplyStroke(withAim(base, -FixedOne/2))

	if right.Velocity.X <= 0 {
		t.Fatalf("right aim lateral velocity = %d, want positive", right.Velocity.X)
	}
	if left.Velocity.X >= 0 {
		t.Fatalf("left aim lateral velocity = %d, want negative", left.Velocity.X)
	}
	if right.AngularVelocity.Y <= 0 {
		t.Fatalf("right aim sidespin = %d, want positive", right.AngularVelocity.Y)
	}
	if left.AngularVelocity.Y >= 0 {
		t.Fatalf("left aim sidespin = %d, want negative", left.AngularVelocity.Y)
	}
	if right.Velocity.Z != left.Velocity.Z {
		t.Fatalf("aim should not change forward power at perfect timing: right=%d left=%d", right.Velocity.Z, left.Velocity.Z)
	}

	clamped := ApplyStroke(withAim(base, FixedOne*2))
	edge := ApplyStroke(withAim(base, FixedOne))
	if clamped != edge {
		t.Fatalf("extreme aim should clamp to +1: got %+v want %+v", clamped, edge)
	}
	clampedLeft := ApplyStroke(withAim(base, -FixedOne*2))
	leftEdge := ApplyStroke(withAim(base, -FixedOne))
	if clampedLeft != leftEdge {
		t.Fatalf("extreme aim should clamp to -1: got %+v want %+v", clampedLeft, leftEdge)
	}
}

func TestApplyStrokeTimingQualityReducesPowerAndAddsAimError(t *testing.T) {
	base := StrokeInput{
		Swing: SwingFlat,
		Aim:   0,
		Power: FixedOne,
	}

	perfect := ApplyStroke(withTiming(base, FixedOne))
	late := ApplyStroke(withTiming(base, FixedOne/2))
	veryLate := ApplyStroke(withTiming(base, 0))

	if !(perfect.Velocity.Z > late.Velocity.Z && late.Velocity.Z > veryLate.Velocity.Z) {
		t.Fatalf("forward speed should decrease with timing quality: perfect=%d late=%d veryLate=%d", perfect.Velocity.Z, late.Velocity.Z, veryLate.Velocity.Z)
	}
	if !(AbsFixed(perfect.Velocity.X) < AbsFixed(late.Velocity.X) && AbsFixed(late.Velocity.X) < AbsFixed(veryLate.Velocity.X)) {
		t.Fatalf("aim error should increase as timing worsens: perfect=%d late=%d veryLate=%d", perfect.Velocity.X, late.Velocity.X, veryLate.Velocity.X)
	}

	replayed := ApplyStroke(withTiming(base, FixedOne/2))
	if replayed != late {
		t.Fatalf("stroke output must be deterministic: got %+v want %+v", replayed, late)
	}
}

func TestApplyStrokeZeroPowerProducesRestState(t *testing.T) {
	output := ApplyStroke(StrokeInput{
		Swing:         SwingSlice,
		Aim:           FixedOne,
		Power:         0,
		TimingQuality: 0,
	})

	if !output.Velocity.IsZero() {
		t.Fatalf("zero power velocity = %+v, want zero", output.Velocity)
	}
	if !output.AngularVelocity.IsZero() {
		t.Fatalf("zero power angular velocity = %+v, want zero", output.AngularVelocity)
	}
}

func withSwing(input StrokeInput, swing SwingType) StrokeInput {
	input.Swing = swing
	return input
}

func withAim(input StrokeInput, aim Fixed) StrokeInput {
	input.Aim = aim
	return input
}

func withTiming(input StrokeInput, timing Fixed) StrokeInput {
	input.TimingQuality = timing
	return input
}
