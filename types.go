package overpass

import "time"

// ElementType represents possible types for Overpass response elements.
type ElementType string

// Possible values are node, way and relation.
const (
	ElementTypeNode     ElementType = "node"
	ElementTypeWay      ElementType = "way"
	ElementTypeRelation ElementType = "relation"
)

// Meta contains fields common for all OSM types.
type Meta struct {
	ID        int64              `json:"id"`
	Timestamp *time.Time         `json:"timestamp,omitempty"`
	Version   int64              `json:"version,omitempty"`
	Changeset int64              `json:"changeset,omitempty"`
	User      string             `json:"user,omitempty"`
	UID       int64              `json:"uid,omitempty"`
	Tags      map[string]string  `json:"tags,omitempty"`
}

// Node represents OSM node type.
type Node struct {
	Meta
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// Way represents OSM way type.
type Way struct {
	Meta
	Nodes    []*Node `json:"nodes,omitempty"`
	Bounds   *Box    `json:"bounds,omitempty"`
	Geometry []Point `json:"geometry,omitempty"`
}

type Point struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// Relation represents OSM relation type.
type Relation struct {
	Meta
	Members []RelationMember `json:"members,omitempty"`
	Bounds  *Box             `json:"bounds,omitempty"`
}

type Box struct {
	Min Point `json:"min"`
	Max Point `json:"max"`
}

// RelationMember represents OSM relation member type.
type RelationMember struct {
	Type     ElementType `json:"type"`
	Node     *Node       `json:"node,omitempty"`
	Way      *Way        `json:"way,omitempty"`
	Relation *Relation   `json:"relation,omitempty"`
	Role     string      `json:"role,omitempty"`
}

// Result returned by Query and contains parsed result of Overpass query.
type Result struct {
	Timestamp time.Time          `json:"timestamp"`
	Count     int                `json:"count"`
	Nodes     map[int64]*Node    `json:"nodes,omitempty"`
	Ways      map[int64]*Way     `json:"ways,omitempty"`
	Relations map[int64]*Relation `json:"relations,omitempty"`
}
