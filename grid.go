package hexgrid
/* 
*  File: hexgrid.go
*  Author: Bryan Matsuo [bmatsuo@soe.ucsc.edu] 
*  Created: Tue Jun 28 03:40:52 PDT 2011
 */
import (
    "fmt"
    "math"
    //"log"
)

var (
    hexHighVertexIncidenceOffset = [][][]int{
        {{-1, 0, 2}, {0, -1, 4}},
        {{0, -1, 3}, {1, 0, 5}},
        {{1, 0, 4}, {1, 1, 0}},
        {{1, 1, 5}, {0, 1, 1}},
        {{0, 1, 0}, {-1, 1, 2}},
        {{-1, 1, 2}, {-1, 0, 3}}}
    hexLowVertexIncidenceOffset = [][][]int{
        {{-1, -1, 2}, {0, -1, 4}},
        {{0, -1, 3}, {1, -1, 5}},
        {{1, -1, 4}, {1, 0, 0}},
        {{1, 0, 5}, {0, 1, 1}},
        {{0, 1, 0}, {-1, 0, 2}},
        {{-1, 0, 2}, {-1, -1, 3}}}
)

//  Discrete hex coordinates consist of a horizontal U axis and a vertical
//  V axis. Each axis has range (-inf,inf) in theory. In practice, Grid
//  objects limit the accessible hex tiles.
type Coords struct {
    U, V int
}
//  For each coordinate in a Grid there is one unique HexTile.
type Tile struct {
    Coords Coords
    Pos    Point
    Value  interface{}
}
//  A HexEdge represents an edge between two HexVertex objects. It is
//  part of the boundary of a HexTile. A HexEdge can be 'Between' only
//  one tile if its tile is on the edge of the grid.
type Edge struct {
    Value interface{}
}
//  A HexVertex represents the corner a HexTile. A HexVertex can be shared
//  by at most 3 HexTiles and can be the junction of between 2 and three
//  HexEdge objects. A HexVertex can be 'between' fewer than three tiles if
//  its tiles are on the edge of the grid. It will be the endpoint of two
//  edges only it belongs to one tile (and is on the edge of the grid).
type Vertex struct {
    Pos   Point
    Value interface{}
}

//  A grid of hexagons in a discrete coordinate system (u,v) where u
//  indexes the column of the grid, and v the row.
type Grid struct {
    radius   float64
    n        int
    p        []Point
    v        []Vertex
    e        []Edge
    hexes    [][]*HexPoints
    tiles    [][]Tile
    vertices [][][]*Vertex
    edges    [][][][]*Edge
}

func (h *Grid) genTiles() {
    // Generate all tiles.
    h.tiles = make([][]Tile, h.n)
    for i := 0; i < h.n; i++ {
        h.tiles[i] = make([]Tile, h.n)
        for j := 0; j < h.n; j++ {
            var (
                u, v   = h.hexCoords(i, j)
                center = h.tileCenter(u, v)
            )
            h.tiles[i][j] = Tile{Coords: Coords{u, v}, Pos: center, Value: 0}
        }
    }
}
func (h *Grid) genVertices() {
    // Make space for vertices/pointers.
    h.v = make([]Vertex, 0, 2*int(math.Pow(float64(h.n), 2)+2*float64(h.n)))
    h.vertices = make([][][]*Vertex, h.n)
    for i := 0; i < h.n; i++ {
        h.vertices[i] = make([][]*Vertex, h.n)
        for j := 0; j < h.n; j++ {
            h.vertices[i][j] = make([]*Vertex, 6)
        }
    }
    // Generate vertices
    for i := 0; i < h.n; i++ {
        for j := 0; j < h.n; j++ {
            var (
                u, v = h.hexCoords(i, j)
                hex  = h.GetHex(u, v)
            )
            for k := 0; k < 6; k++ {
                if h.vertices[i][j][k] == nil {
                    var identVertices = h.GetVerticesIdentical(u, v, k)
                    if identVertices == nil {
                        panic("outofbounds")
                    }
                    h.v = append(h.v, Vertex{Pos: hex[k], Value: 0})
                    for _, ident := range identVertices {
                        var (
                            uIdent         = ident[0]
                            vIdent         = ident[1]
                            kIdent         = ident[2]
                            iIdent, jIdent = h.hexIndex(uIdent, vIdent)
                        )
                        if h.WithinBounds(uIdent, vIdent) {
                            h.vertices[iIdent][jIdent][kIdent] = &(h.v[len(h.v)-1])
                        }
                    }
                }
            }
        }
    }
}
func (h *Grid) genEdges() {
    // Make space for edges/pointers.
    h.e = make([]Edge, 0, 3*int(math.Pow(float64(h.n), 2))+4*h.n-1)
    h.edges = make([][][][]*Edge, h.n)
    for i := 0; i < h.n; i++ {
        h.edges[i] = make([][][]*Edge, h.n)
        for j := 0; j < h.n; j++ {
            h.edges[i][j] = make([][]*Edge, 6)
            for k := 0; k < 6; k++ {
                h.edges[i][j][k] = make([]*Edge, 6)
            }
        }
    }
    // Generate all edges.
    for i := 0; i < h.n; i++ {
        for j := 0; j < h.n; j++ { // BEGIN (u,v) TILE ANALYSIS
            var (
                u, v = h.hexCoords(i, j)
            )
            for k := 0; k < 6; k++ {
                for ell := 0; ell < 6; ell++ { // BEGIN (k,ell) EDGE ANALYSIS
                    // Ensure an edge between k and ell exists.
                    var (
                        hexTmp  = &HexPoints{}
                        edgeDir = hexTmp.EdgeDirection(k, ell)
                    )
                    if edgeDir != NilDirection && h.edges[i][j][k][ell] == nil {
                        // Create the edge, compute the other incident tile.
                        h.e = append(h.e, Edge{Value: 0})
                        var (
                            edgePtr        = &(h.e[len(h.e)-1])
                            adjEdgeIndices = hexTmp.EdgeIndices(edgeDir.Inverse())
                        )
                        if adjEdgeIndices == nil {
                            panic("niladjindices")
                        }
                        var (
                            adjK   = adjEdgeIndices[0]
                            adjEll = adjEdgeIndices[1]
                        )
                        if edgeDir == NilDirection {
                            panic("niledgedirection")
                        }
                        var adjCoordsSlice = h.GetHexAdjacentCoords(u, v, edgeDir)
                        if adjCoordsSlice == nil {
                            panic("niladjcoords")
                        }
                        // Store the edge pointer is its various configurations.
                        var (
                            adjCoords  = adjCoordsSlice[0]
                            adjU       = adjCoords[0]
                            adjV       = adjCoords[1]
                            adjI, adjJ = h.hexIndex(adjU, adjV)
                        )
                        if h.WithinBounds(adjU, adjV) {
                            h.edges[adjI][adjJ][adjK][adjEll] = edgePtr
                            h.edges[adjI][adjJ][adjEll][adjK] = edgePtr
                        }
                        h.edges[i][j][k][ell] = edgePtr
                        h.edges[i][j][ell][k] = edgePtr
                    }
                }   // END (k,ell) EDGE ANALYSIS
            }
        }   // END (u,v) TILE ANALYSIS
    }
}
func (h *Grid) genHexagons() {
    h.p = make([]Point, 0, 2*int(math.Pow(float64(h.n), 2)+2*float64(h.n)))
    h.hexes = make([][]*HexPoints, h.n)
    var indexOffset = h.indexOffset()

    // Generate all hexagons.
    for i := 0; i < h.n; i++ {
        h.hexes[i] = make([]*HexPoints, h.n)
        for j := 0; j < h.n; j++ {
            var (
                u, v = h.hexCoords(i, j)
            )
            h.hexes[i][j] = h.GetHex(u, v)
            if h.hexes[i][j] == nil {
                panic(fmt.Sprintf("OutOfBounds(%d,%d)", i, j))
            }
        }
    }

    // Collect all points, sharing common points belonging to adjacent hexagons.
    for i := 0; i < h.n; i++ {
        for j := 0; j < h.n; j++ {
            var (
                toAdd = [6]bool{true, true, true, true, true, true}
                u     = i - indexOffset
                hex   = h.hexes[i][j]
            )
            var iHigh = h.columnIsHigh(u)
            // Account for the points in the hex tile above i,j
            if i > 0 {
                if iHigh {
                    hex[5] = h.hexes[i-1][j][3]
                    toAdd[5] = false
                    hex[0] = h.hexes[i-1][j][2]
                    toAdd[0] = false
                } else {
                    hex[4] = h.hexes[i-1][j][2]
                    toAdd[4] = false
                    hex[5] = h.hexes[i-1][j][1]
                    toAdd[5] = false
                    if j > 0 {
                        hex[5] = h.hexes[i-1][j-1][3]
                        toAdd[5] = false
                    }
                }
            }

            // Account for points in the hex tile above and to the left of i,j
            if j > 0 {
                hex[0] = h.hexes[i][j-1][4]
                toAdd[0] = false
                hex[1] = h.hexes[i][j-1][3]
                toAdd[1] = false
            }

            // Account for corner cases.
            if !iHigh && i < h.n-1 && j > 0 {
                hex[2] = h.hexes[i+1][j-1][4]
                toAdd[2] = false
            }

            // Account for points in the hex tile not already accounted for.
            for k, shouldAdd := range toAdd {
                if !shouldAdd {
                    continue
                }
                h.p = append(h.p, hex[k])
                if h.p == nil {
                    panic("nil append result")
                }
            }
        }
    }
}

//  Create an nxn grid of hexagons with radius r.
//  The integer n must be odd.
func NewGrid(n int, r float64) *Grid {
    if n&1 == 0 {
        panic("evensize")
    }
    if n < 0 {
        panic("negsize")
    }
    if r < 0 {
        panic("negradius")
    }
    var h = new(Grid)
    h.radius = r
    h.n = n
    h.genHexagons()
    h.genTiles()
    h.genVertices()
    h.genEdges() // Must come after genVertices.
    return h
}

//  Length of a single dimension in the field (n).
func (h *Grid) Size() int {
    return h.n
}
//  Total number of distinct hexagon vertices in the field.
func (h *Grid) NumPoints() int {
    return len(h.p)
}
//  Number of hex tiles in the field (n^2).
func (h *Grid) NumTiles() int {
    return h.n * h.n
}
func (h *Grid) indexOffset() int {
    return int(math.Floor(float64(h.n) / 2))
}
//  Minimum value of the row coordinate v.
func (h *Grid) RowMin() int { return -h.indexOffset() }
//  Maximum value of the row coordinate v.
func (h *Grid) RowMax() int { return h.indexOffset() }
//  Minimum value of the column coordinate u.
func (h *Grid) ColMin() int { return -h.indexOffset() }
//  Maximum value of the column coordinate u.
func (h *Grid) ColMax() int { return h.indexOffset() }


/* Some coordinate <-> index internal methods. */
func (h *Grid) hexCoords(i, j int) (int, int) {
    var offset = h.indexOffset()
    return i - offset, j - offset
}
func (h *Grid) hexIndex(u, v int) (int, int) {
    var offset = h.indexOffset()
    return u + offset, v + offset
}

/* Internal bounds checking method. */
func (h *Grid) indexWithinBounds(i, j int) bool {
    u, v := h.hexCoords(i, j)
    return h.WithinBounds(u, v)
}
//  Returns true if the hex at coordinates (u,v) is in the hex field.
func (h *Grid) WithinBounds(u, v int) bool {
    var offset = h.indexOffset()
    if int(math.Fabs(float64(u))) > offset {
        return false
    }
    if int(math.Fabs(float64(v))) > offset {
        return false
    }
    return true
}

//  Generate points for the hexagon at row i, column j.
//  Returns nil when the position (i,j) is not within the bounds of the board.
func (h *Grid) GetHex(u, v int) *HexPoints {
    if !h.WithinBounds(u, v) {
        return nil
    }
    var (
        i, j = h.hexIndex(u, v)
    )
    if h.hexes[i][j] != nil {
        var newh = new(HexPoints)
        *newh = *(h.hexes[i][j])
        return newh
    }
    return NewHex(h.tileCenter(u, v), h.radius)
}

//  Get a pointer to the kth corner point of the hex tile at (u,v).
//  Returns PointInf() when (u,v) is not within the bounds of h.
func (h *Grid) GetVertexPoint(u, v, k int) Point {
    var hex = h.GetHex(u, v)
    if hex == nil {
        return PointInf()
    }
    return hex[k]
}


//  Determine if (u1,v1,k1) and (u2,v2,k2) reference the same point.
func (h *Grid) VerticesAreIdentical(u1, v1, k1, u2, v2, k2 int) bool {
    var identVertices = h.GetVerticesIdentical(u1, v1, k1)
    if identVertices == nil {
        panic("nilident")
    }
    for _, ident := range identVertices {
        if ident[0] == u2 && ident[1] == v2 && ident[2] == k2 {
            return true
        }
    }
    return false
}

//  Get coordinates of hex vertices in the field incident to vertex (u,v,k).
//  Returns a slice of vertex coordinates (slices of 3 ints), the first of
//  which being []int{u, v, k}. See also, VerticesAreIdentical.
func (h *Grid) GetVerticesIdentical(u, v, k int) [][]int {
    var adjC = make([][]int, 0, 3)
    adjC = append(adjC, []int{u, v, k})

    var adjOffsets [][]int
    if h.columnIsHigh(u) {
        adjOffsets = hexHighVertexIncidenceOffset[k]
    } else {
        adjOffsets = hexLowVertexIncidenceOffset[k]
    }
    for _, offset := range adjOffsets {
        var (
            du         = offset[0]
            dv         = offset[1]
            kAdj       = offset[2]
            uNew       = u + du
            vNew       = v + dv
            shouldCopy = h.WithinBounds(uNew, vNew)
        )
        if !shouldCopy {
            continue
        }
        var offsetCp = []int{uNew, vNew, kAdj}
        adjC = append(adjC, offsetCp)
    }

    return adjC
}

//  Get the index of the vertex clockwise of vertex k.
func (h *Grid) GetVertexAdjacentClockwise(k int) int {
    return (k + 5) % 6
}
//  Get the index of the vertex counter-clockwise of vertex k.
func (h *Grid) GetVertexAdjacentCounterClockwise(k int) int {
    return (k + 1) % 6
}

//  Get a list of unique vertices adjacent to (u,v,k).
//  See also, VerticesAreIdentical.
func (h *Grid) GetVerticesAdjacent(u, v, k int) [][]int {
    var identVerts = h.GetVerticesIdentical(u, v, k)
    if identVerts == nil {
        return nil
    }
    var adjVerts = make([][]int, len(identVerts))
    for i, vert := range identVerts {
        adjVerts[i] = []int{vert[0], vert[1], h.GetVertexAdjacentClockwise(vert[2])}
    }
    return adjVerts
}

//  Get hex tiles incident with the kth corner point of hex at (u,v).
//  Returns nil when (u,v) is not within the bounds of h.
//  Otherwise, a slice of *HexPoints is returned w/ hex tile (u,v) at index 0.
func (h *Grid) GetHexIncident(u, v, k int) []*HexPoints {
    var hex = h.GetHex(u, v)
    if hex == nil {
        return nil
    }
    var adjC = h.GetVerticesIdentical(u, v, k)
    var adj = make([]*HexPoints, 0, len(adjC))
    for _, coords := range adjC {
        var (
            hexAdj = h.GetHex(coords[0], coords[1])
        )
        if hexAdj == nil {
            panic("coordoutofbounds")
        }
        adj = append(adj, hexAdj)
    }

    return adj
}

/* Internal methods for computing hexagon positions. */
func (h *Grid) columnIsHigh(u int) bool {
    var (
        offset     = uint(h.indexOffset())
        i          = uint(u + int(offset))
        iOdd       = i % 2
        sideIsHigh = offset % 2
    )
    return iOdd^sideIsHigh == 1
}
func (h *Grid) horizontalSpacing() float64 {
    return 2 * h.radius * math.Cos(hexTriangleAngle)
}
func (h *Grid) verticalSpacing() float64 {
    return 2 * h.radius
}
func (h *Grid) verticalOffset(u int) float64 {
    if h.columnIsHigh(u) {
        return 2 * h.radius * math.Sin(hexTriangleAngle)
    }
    return 0
}

func (h *Grid) tileCenter(u, v int) Point {
    var (
        centerX = float64(u) * h.horizontalSpacing()
        centerY = float64(v) * h.verticalSpacing()
    )
    centerY += h.verticalOffset(u)
    return Point{centerX, centerY}
}


//  Return the direction of vertex k relative to the center of a hexagon.
//  Returns NilDirection if k is not in the range [0,5].
func (h *Grid) VertexDirection(k int) HexDirection {
    switch k {
    case 0:
        return SW
    case 1:
        return SE
    case 2:
        return E
    case 3:
        return NE
    case 4:
        return NW
    case 5:
        return W
    }
    return NilDirection
}
//  Return the vertex k in direction dir from a hex tile's center.
//  Returns -1 if dir is NilDirection, N, or S.
func (h *Grid) VertexIndex(dir HexDirection) int {
    switch dir {
    case SW:
        return 0
    case SE:
        return 1
    case E:
        return 2
    case NE:
        return 3
    case NW:
        return 4
    case W:
        return 5
    }
    return -1
}

//  Return a slice of hexagons adjacent to the hex tile at coordinates (u, v).
//  Only hex tiles in the Grid are returned.
//  If (u,v) is not in the Grid, a nil slice is returned.
func (h *Grid) GetHexAdjacent(u, v int, dir HexDirection) []*HexPoints {
    if !h.WithinBounds(u, v) {
        return nil
    }
    var adjC = h.GetHexAdjacentCoords(u, v, dir)
    if adjC == nil {
        panic("niladjacency")
    }
    var adj = make([]*HexPoints, 0, len(adjC))
    for _, coords := range adjC {
        var (
            uPrime = coords[0]
            vPrime = coords[1]
        )
        if !h.WithinBounds(uPrime, vPrime) {
            continue
        }
        adj = append(adj, h.GetHex(uPrime, vPrime))
    }
    return adj
}

//  Return a slice of the coordinates for adjacent hexagons
//  (not necessarily in the grid).
//  If E (or W) is supplied then the NE and SE (or NE and SW) coordinates
//  are returned in that order.
//  If NilDirection is suppied, then coordinates for all adjacent hexagons
//  are returned in the order N, NE, SE, S, SW, NW.
func (h *Grid) GetHexAdjacentCoords(u, v int, dir HexDirection) [][]int {
    switch dir {
    case N:
        return [][]int{{u, v + 1}}
    case S:
        return [][]int{{u, v - 1}}
    case E:
        var adjE = make([][]int, 2)
        if h.columnIsHigh(u) {
            adjE[0] = []int{u - 1, v + 1}
            adjE[1] = []int{u - 1, v}
        } else {
            adjE[0] = []int{u - 1, v}
            adjE[1] = []int{u - 1, v - 1}
        }
        return adjE
    case W:
        var adjW = make([][]int, 2)
        if h.columnIsHigh(u) {
            adjW[0] = []int{u + 1, v + 1}
            adjW[1] = []int{u + 1, v}
        } else {
            adjW[0] = []int{u + 1, v}
            adjW[1] = []int{u + 1, v - 1}
        }
        return adjW
    case NE:
        if h.columnIsHigh(u) {
            return [][]int{{u - 1, v + 1}}
        }
        return [][]int{{u - 1, v}}
    case NW:
        if h.columnIsHigh(u) {
            return [][]int{{u + 1, v + 1}}
        }
        return [][]int{{u + 1, v}}
    case SE:
        if h.columnIsHigh(u) {
            return [][]int{{u - 1, v}}
        }
        return [][]int{{u - 1, v - 1}}
    case SW:
        if h.columnIsHigh(u) {
            return [][]int{{u + 1, v}}
        }
        return [][]int{{u + 1, v - 1}}
    default:
        var adjAll = make([][]int, 6)
        if h.columnIsHigh(u) {
            adjAll[0] = []int{u, v + 1}     // North
            adjAll[1] = []int{u - 1, v + 1} // NorthEast
            adjAll[2] = []int{u - 1, v}     // SouthEast
            adjAll[3] = []int{u, v - 1}     // South
            adjAll[4] = []int{u + 1, v}     // SouthWest
            adjAll[5] = []int{u + 1, v + 1} // NorthWest
        } else {
            adjAll[0] = []int{u, v + 1}
            adjAll[1] = []int{u - 1, v}
            adjAll[2] = []int{u - 1, v - 1}
            adjAll[3] = []int{u, v - 1}
            adjAll[4] = []int{u + 1, v - 1}
            adjAll[5] = []int{u + 1, v}
        }
        return adjAll
    }
    return nil
}

//  If hex tiles (u1,v1) and (u2,v2) are adjacent, the direction of (u2,v2)
//  from (u1,v1) is returned. Otherwise NilDirection is returned.
func (h *Grid) HexAdjacency(u1, v1, u2, v2 int) HexDirection {
    var (
        deltaU = u2 - u1
        deltaV = v2 - v1
    )
    if u1 == u2 {
        if deltaV == 1 {
            return N
        } else if deltaV == -1 {
            return S
        }
        return NilDirection
    }
    if deltaU == 1 {
        if h.columnIsHigh(u1) {
            if v1 == v2 {
                return SE
            }
            if v2 == v1+1 {
                return NE
            }
        } else {
            if v2 == v1-1 {
                return SE
            }
            if v1 == v2 {
                return NE
            }
        }
    } else if deltaU == -1 {
        if h.columnIsHigh(u1) {
            if v2 == v1+1 {
                return NW
            }
            if v1 == v2 {
                return SW
            }
        } else {
            if v1 == v2 {
                return NW
            }
            if v2 == v1-1 {
                return SW
            }
        }
    }
    return NilDirection
}

//  Function for determining the vertex indices of any edge in
//  the hex tile at (u1,v1) that is alse in tile (u2,v2).
//  Returns nil if the hex coordinates are not adjacent.
//  See also, SharedEdge
func (h *Grid) SharedEdgeIndices(u1, v1, u2, v2 int) []int {
    var (
        adjDir = h.HexAdjacency(u1, v1, u2, v2)
        tmpHex = HexPoints{}
    )
    if adjDir == NilDirection {
        return nil
    }
    return tmpHex.EdgeIndices(adjDir)
}

//  Function for determining the actual points determining any shared edge
//  between hex tiles (u1,v1) and (u2,v2).
//  Returns nil if either coordinates are outside of the hex field.
//  Returns nil if the hex coordinates are not adjacent.
//  See also, SharedEdgeIndices
func (h *Grid) SharedEdge(u1, v1, u2, v2 int) []Point {
    if !h.WithinBounds(u1, v1) {
        return nil
    }
    if !h.WithinBounds(u2, v2) {
        return nil
    }
    var sharedIndices = h.SharedEdgeIndices(u1, v1, u2, v2)
    if sharedIndices == nil {
        return nil
    }
    var h1 = h.GetHex(u1, v1)
    return []Point{h1[sharedIndices[0]], h1[sharedIndices[1]]}
}
