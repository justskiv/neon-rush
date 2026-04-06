package main

// Rect is an axis-aligned bounding box (top-left origin).
type Rect struct {
	X, Y, W, H float64
}

// NewRect creates a Rect from center coordinates.
func NewRect(cx, cy, w, h float64) Rect {
	return Rect{
		X: cx - w/2,
		Y: cy - h/2,
		W: w,
		H: h,
	}
}

// CheckCollision returns true if two AABBs overlap.
func CheckCollision(a, b Rect) bool {
	return a.X < b.X+b.W &&
		a.X+a.W > b.X &&
		a.Y < b.Y+b.H &&
		a.Y+a.H > b.Y
}

// CollisionResult holds collision detection output.
type CollisionResult struct {
	Hit     bool
	HitX    float64
	HitY    float64
	CarType CarType
}

// CheckPlayerTrafficCollision checks if the player collides with any NPC.
func CheckPlayerTrafficCollision(p *Player, traffic []*TrafficCar) CollisionResult {
	pb := p.Bounds()
	for _, c := range traffic {
		if CheckCollision(pb, c.Bounds()) {
			return CollisionResult{
				Hit:     true,
				HitX:    (p.X + c.X) / 2,
				HitY:    (p.Y + c.Y) / 2,
				CarType: c.CarType,
			}
		}
	}
	return CollisionResult{}
}
