package render

import "github.com/plexusone/systemspec-architecture/sas"

// Shape is an abstract node shape, mapped to concrete syntax by each
// format-specific renderer, so every format visually distinguishes the
// same node kinds the same way.
type Shape string

const (
	ShapeRectangle Shape = "rectangle"
	ShapeCylinder  Shape = "cylinder"
	ShapeStadium   Shape = "stadium"
	ShapeHexagon   Shape = "hexagon"
)

// ShapeForKind chooses a Shape for a NodeKind: data.database renders as a
// cylinder, actor as a stadium (person-adjacent shape), external_service
// as a hexagon (visually distinct from the system's own components), and
// everything else as a rectangle.
func ShapeForKind(kind sas.NodeKind) Shape {
	switch kind {
	case sas.NodeKindDataDatabase:
		return ShapeCylinder
	case sas.NodeKindActor:
		return ShapeStadium
	case sas.NodeKindExternalService:
		return ShapeHexagon
	default:
		return ShapeRectangle
	}
}
