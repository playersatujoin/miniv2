package world

// Layer identifies which map layer a tile belongs to.
type Layer string

const (
	LayerGround  Layer = "ground"
	LayerObjects Layer = "objects"
)

// TileDef describes a tile so the client knows how to draw and collide with it.
type TileDef struct {
	ID    int    `json:"id"`
	Key   string `json:"key"`
	Name  string `json:"name"`
	Layer Layer  `json:"layer"`
	Color string `json:"color"`
	Solid bool   `json:"solid"`
}

// Ground tile IDs.
const (
	DeepWater = iota
	Water
	Sand
	Grass
	ForestFloor
	Dirt
	StoneFloor
	Mountain
	Bridge
	VolcanicRock
	Crater
	Limestone
)

// Object tile IDs. None means the cell is empty.
const (
	None = iota
	Tree
	Pine
	Boulder
	Bush
	Flowers
	Wall
)

var GroundTiles = []TileDef{
	{DeepWater, "deep_water", "Air Dalam", LayerGround, "#1f4e8c", true},
	{Water, "water", "Air", LayerGround, "#2f74c0", true},
	{Sand, "sand", "Pasir", LayerGround, "#e2cf8f", false},
	{Grass, "grass", "Rumput", LayerGround, "#6aab4f", false},
	{ForestFloor, "forest_floor", "Lantai Hutan", LayerGround, "#4b8a3d", false},
	{Dirt, "dirt", "Jalan Tanah", LayerGround, "#a07a4f", false},
	{StoneFloor, "stone_floor", "Lantai Batu", LayerGround, "#8d8a83", false},
	{Mountain, "mountain", "Gunung", LayerGround, "#5b5650", true},
	{Bridge, "bridge", "Jembatan", LayerGround, "#9b6b3c", false},
	{VolcanicRock, "batuan_vulkanik", "Batuan Vulkanik", LayerGround, "#4a4340", false},
	{Crater, "kawah", "Kawah", LayerGround, "#3fb7b0", true},
	{Limestone, "batu_gamping", "Batu Gamping", LayerGround, "#cfcabb", false},
}

var ObjectTiles = []TileDef{
	{None, "none", "Penghapus", LayerObjects, "#00000000", false},
	{Tree, "tree", "Pohon Ek", LayerObjects, "#2f6b2a", true},
	{Pine, "pine", "Pohon Pinus", LayerObjects, "#24543a", true},
	{Boulder, "rock", "Batu", LayerObjects, "#7c7c7c", true},
	{Bush, "bush", "Semak", LayerObjects, "#3f8f3a", false},
	{Flowers, "flowers", "Bunga", LayerObjects, "#e98ad0", false},
	{Wall, "wall", "Tembok Batu", LayerObjects, "#6b625a", true},
}

func validID(defs []TileDef, id int) bool { return id >= 0 && id < len(defs) }

func groundSolid(id int) bool { return GroundTiles[id].Solid }
func objectSolid(id int) bool { return ObjectTiles[id].Solid }
