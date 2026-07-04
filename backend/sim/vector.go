package sim

type Vector3 struct {
	X Fixed
	Y Fixed
	Z Fixed
}

func (v Vector3) Scale(s Fixed) Vector3 {
	return Vector3{
		X: v.X.Mul(s),
		Y: v.Y.Mul(s),
		Z: v.Z.Mul(s),
	}
}

func (v Vector3) Add(other Vector3) Vector3 {
	return Vector3{
		X: v.X + other.X,
		Y: v.Y + other.Y,
		Z: v.Z + other.Z,
	}
}

func (v Vector3) IsZero() bool {
	return v.X == 0 && v.Y == 0 && v.Z == 0
}
