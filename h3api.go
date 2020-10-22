package h3

import (
	"bytes"
	"fmt"
)

// Maximum number of cell boundary vertices; worst case is pentagon:
// 5 original Verts + 5 edge crossings
const MAX_CELL_BNDRY_VERTS = 10

/**
  @brief latitude/longitude in radians
*/
type geoCoord struct {
	Lat float64 ///< latitude in radians
	Lon float64 ///< longitude in radians
}

// GeoFromDegrees creates geoCoords from degrees
func GeoFromDegrees(lat, lon float64) *geoCoord {
	return GeoFromRadians(lat, lon).AsRadians()
}

// GeoFromRadians creates geoCoords from Radians
func GeoFromRadians(lat, lon float64) *geoCoord {
	return &geoCoord{
		Lat: lat,
		Lon: lon,
	}
}

func (g geoCoord) String() string {
	return fmt.Sprintf("%f,%f", g.Lat, g.Lon)
}

// AsDegrees converts the coordinates from radians to degrees.
func (g geoCoord) AsDegrees() *geoCoord {
	return &geoCoord{
		Lon: normalizeDegree(radsToDegs(g.Lon), -180.0, 180),
		Lat: normalizeDegree(radsToDegs(g.Lat), -90, 90),
	}
}

// AsRadians converts the coordinates from degrees to radians.
func (g geoCoord) AsRadians() *geoCoord {
	return &geoCoord{
		Lon: degsToRads(g.Lon),
		Lat: degsToRads(g.Lat),
	}
}

/**
  @brief cell boundary in latitude/longitude
*/
type GeoBoundary struct {
	numVerts int        ///< number of vertices
	Verts    []geoCoord ///< vertices in ccw order
}

func (gb GeoBoundary) String() string {
	buf := bytes.NewBuffer(nil)
	buf.WriteRune('[')
	for i := range gb.Verts {
		if i != 0 {
			buf.WriteRune(' ')
		}
		buf.WriteString(gb.Verts[i].String())
	}
	buf.WriteRune(']')
	return buf.String()
}

func (gb GeoBoundary) AsDegrees() *GeoBoundary {
	list := make([]geoCoord, len(gb.Verts))

	for i := range gb.Verts {
		list[i] = *(gb.Verts[i].AsDegrees())
	}

	gb.Verts = list

	return &gb
}

func (gb GeoBoundary) AsRadians() *GeoBoundary {
	list := make([]geoCoord, len(gb.Verts))

	for i := range gb.Verts {
		list[i] = *(gb.Verts[i].AsDegrees())
	}

	gb.Verts = list

	return &gb
}

/**
 *  @brief similar to GeoBoundary, but requires more alloc work
 */
type Geofence struct {
	numVerts int
	verts    []geoCoord
}

func (g *Geofence) IsZero() bool {
	return g == nil || g.numVerts == 0
}

func (g *Geofence) NewIterate() func(vertexA *geoCoord, vertexB *geoCoord) bool {
	loopIndex := -1

	return func(vertexA *geoCoord, vertexB *geoCoord) bool {
		loopIndex++

		if loopIndex >= g.numVerts {
			return false
		}

		*vertexA = g.verts[loopIndex]
		*vertexB = g.verts[(loopIndex+1)%g.numVerts]

		return true
	}
}

/**
 *  @brief Simplified core of GeoJSON Polygon coordinates definition
 */
type GeoPolygon struct {
	geofence Geofence   ///< exterior boundary of the polygon
	numHoles int        ///< number of elements in the array pointed to by holes
	holes    []Geofence ///< interior boundaries (holes) in the polygon
}

/**
 *  @brief Simplified core of GeoJSON MultiPolygon coordinates definition
 */
type GeoMultiPolygon struct {
	numPolygons int
	polygons    []GeoPolygon
}

/**
 *  @brief A coordinate node in a linked geo structure, part of a linked list
 */
type LinkedgeoCoord struct {
	vertex geoCoord
	next   *LinkedgeoCoord
}

/**
 *  @brief A loop node in a linked geo structure, part of a linked list
 */
type LinkedGeoLoop struct {
	first *LinkedgeoCoord
	last  *LinkedgeoCoord
	next  *LinkedGeoLoop
}

func (l *LinkedGeoLoop) IsZero() bool {
	return l == nil || l.first == nil
}

func (l *LinkedGeoLoop) NewIterate() func(vertexA *geoCoord, vertexB *geoCoord) bool {
	var currentCoord, nextCoord *LinkedgeoCoord

	return func(vertexA *geoCoord, vertexB *geoCoord) bool {
		var getNextCoord = func(coordToCheck *LinkedgeoCoord) *LinkedgeoCoord {
			if coordToCheck == nil {
				return l.first
			}
			return currentCoord.next

		}

		currentCoord = getNextCoord(currentCoord)

		if currentCoord == nil {
			return false
		}

		*vertexA = currentCoord.vertex
		nextCoord = getNextCoord(currentCoord.next)
		*vertexB = nextCoord.vertex

		return true
	}
}

/**
 *  @brief A polygon node in a linked geo structure, part of a linked list.
 */
type LinkedGeoPolygon struct {
	first *LinkedGeoLoop
	last  *LinkedGeoLoop
	next  *LinkedGeoPolygon
}

/**
 * @brief IJ hexagon coordinates
 *
 * Each axis is spaced 120 degrees apart.
 */
type CoordIJ struct {
	i int ///< i component
	j int ///< j component
}
