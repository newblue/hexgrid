package hexgrid
/* 
*  File: polygon.go
*  Author: Bryan Matsuo [bmatsuo@soe.ucsc.edu] 
*  Created: Wed Jun 29 13:56:22 PDT 2011
 */
import (
    "math"
    //"log"
)

//  Hexagons have faces in directions NW, N, NE, SE, S, SW
//  and vertices in directions W, NW, NE, E, SE, SW.
type HexDirection int

const (
    N   HexDirection = iota
    NE
    NW
    S
    SE
    SW
    E
    W
    NilDirection
)

func (dir HexDirection) Inverse() HexDirection {
    switch dir {
    case N:
        return S
    case NE:
        return SW
    case E:
        return W
    case SE:
        return NW
    case S:
        return N
    case SW:
        return NE
    case W:
        return E
    case NW:
        return SE
    }
    return NilDirection
}

const (
    hexTriangleAngle = math.Pi / 6
    hexRotateAngle   = math.Pi / 3
)

var (
    hexSideRadiusRatio = math.Tan(hexTriangleAngle)
)

//  A simple hexagon type thinly wrapping a Point array.
type HexPoints [6]Point

func (hex *HexPoints) Point(k int) Point {
    if k < 0 || k >= len(hex) {
        panic("Point index out of bounds")
    }
    return hex[k]
}
func (hex *HexPoints) Points() []Point {
    var points = make([]Point, 6)
    copy(points, hex[:])
    return points
}
func (hex *HexPoints) EdgeDirection(k, ell int) HexDirection {
    if k > ell {
        var tmp = k
        k = ell
        ell = tmp
    }
    if k == 0 && ell == 1 {
        return S
    } else if k == 1 && ell == 2 {
        return SE
    } else if k == 2 && ell == 3 {
        return NE
    } else if k == 3 && ell == 4 {
        return N
    } else if k == 4 && ell == 5 {
        return NW
    } else if k == 0 && ell == 5 {
        return SW
    }
    return NilDirection
}
func (hex *HexPoints) EdgeIndices(dir HexDirection) []int {
    switch dir {
    case S:
        return []int{0, 1}
    case SE:
        return []int{1, 2}
    case NE:
        return []int{2, 3}
    case N:
        return []int{3, 4}
    case NW:
        return []int{4, 5}
    case SW:
        return []int{5, 0}
    }
    return nil
}
func (hex *HexPoints) Edge(dir HexDirection) []Point {
    var edgeIndices = hex.EdgeIndices(dir)
    if edgeIndices == nil {
        return nil
    }
    var (
        p1  = hex[edgeIndices[0]]
        p2  = hex[edgeIndices[1]]
    )
    return []Point{p1, p2}
}

//  Generate a hexagon at a given point.
func NewHex(p Point, r float64) *HexPoints {
    var (
        hex  = new(HexPoints)
        side = Point{r, 0}.Scale(hexSideRadiusRatio)
    )
    hex[0] = Point{0, -r}.Sub(side)
    for i := 1; i < 6; i++ {
        hex[i] = hex[i-1].Rot(hexRotateAngle)
    }
    for i := 0; i < 6; i++ {
        hex[i] = hex[i].Add(p)
    }
    return hex
}
