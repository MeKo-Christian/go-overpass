package overpass

import (
	"testing"
)

func TestResult_GetNode(t *testing.T) {
	t.Parallel()

	result := Result{
		Nodes: make(map[int64]*Node),
	}

	// Get node that doesn't exist - should create it
	node1 := result.getNode(123)
	if node1 == nil {
		t.Fatal("expected non-nil node")
	}

	if node1.ID != 123 {
		t.Errorf("expected ID 123, got %d", node1.ID)
	}

	// Verify it was added to the map
	if len(result.Nodes) != 1 {
		t.Errorf("expected 1 node in map, got %d", len(result.Nodes))
	}

	// Get same node again - should return existing one
	node2 := result.getNode(123)
	if node2 != node1 {
		t.Error("expected same node instance")
	}

	// Verify map still has only one entry
	if len(result.Nodes) != 1 {
		t.Errorf("expected still 1 node in map, got %d", len(result.Nodes))
	}
}

func TestResult_GetNode_Existing(t *testing.T) {
	t.Parallel()

	// Pre-populate a node
	existingNode := &Node{
		Meta: Meta{ID: 456},
		Lat:  1.0,
		Lon:  2.0,
	}

	result := Result{
		Nodes: map[int64]*Node{
			456: existingNode,
		},
	}

	// Get existing node
	node := result.getNode(456)
	if node != existingNode {
		t.Error("expected existing node instance")
	}

	if node.Lat != 1.0 || node.Lon != 2.0 {
		t.Error("existing node data should be preserved")
	}
}

func TestResult_GetWay(t *testing.T) {
	t.Parallel()

	result := Result{
		Ways: make(map[int64]*Way),
	}

	// Get way that doesn't exist - should create it
	way1 := result.getWay(789)
	if way1 == nil {
		t.Fatal("expected non-nil way")
	}

	if way1.ID != 789 {
		t.Errorf("expected ID 789, got %d", way1.ID)
	}

	// Verify it was added to the map
	if len(result.Ways) != 1 {
		t.Errorf("expected 1 way in map, got %d", len(result.Ways))
	}

	// Get same way again - should return existing one
	way2 := result.getWay(789)
	if way2 != way1 {
		t.Error("expected same way instance")
	}

	// Verify map still has only one entry
	if len(result.Ways) != 1 {
		t.Errorf("expected still 1 way in map, got %d", len(result.Ways))
	}
}

func TestResult_GetWay_Existing(t *testing.T) {
	t.Parallel()

	// Pre-populate a way
	existingWay := &Way{
		Meta:  Meta{ID: 999},
		Nodes: []*Node{{Meta: Meta{ID: 1}}},
	}

	result := Result{
		Ways: map[int64]*Way{
			999: existingWay,
		},
	}

	// Get existing way
	way := result.getWay(999)
	if way != existingWay {
		t.Error("expected existing way instance")
	}

	if len(way.Nodes) != 1 {
		t.Error("existing way data should be preserved")
	}
}

func TestResult_GetRelation(t *testing.T) {
	t.Parallel()

	result := Result{
		Relations: make(map[int64]*Relation),
	}

	// Get relation that doesn't exist - should create it
	rel1 := result.getRelation(111)
	if rel1 == nil {
		t.Fatal("expected non-nil relation")
	}

	if rel1.ID != 111 {
		t.Errorf("expected ID 111, got %d", rel1.ID)
	}

	// Verify it was added to the map
	if len(result.Relations) != 1 {
		t.Errorf("expected 1 relation in map, got %d", len(result.Relations))
	}

	// Get same relation again - should return existing one
	rel2 := result.getRelation(111)
	if rel2 != rel1 {
		t.Error("expected same relation instance")
	}

	// Verify map still has only one entry
	if len(result.Relations) != 1 {
		t.Errorf("expected still 1 relation in map, got %d", len(result.Relations))
	}
}

func TestResult_GetRelation_Existing(t *testing.T) {
	t.Parallel()

	// Pre-populate a relation
	existingRelation := &Relation{
		Meta:    Meta{ID: 222},
		Members: []RelationMember{{Type: ElementTypeNode}},
	}

	result := Result{
		Relations: map[int64]*Relation{
			222: existingRelation,
		},
	}

	// Get existing relation
	rel := result.getRelation(222)
	if rel != existingRelation {
		t.Error("expected existing relation instance")
	}

	if len(rel.Members) != 1 {
		t.Error("existing relation data should be preserved")
	}
}

func TestResult_GetMultiple(t *testing.T) {
	t.Parallel()

	// Test getting multiple different types in one result
	result := Result{
		Nodes:     make(map[int64]*Node),
		Ways:      make(map[int64]*Way),
		Relations: make(map[int64]*Relation),
	}

	node := result.getNode(1)
	way := result.getWay(2)
	rel := result.getRelation(3)

	if node == nil || way == nil || rel == nil {
		t.Fatal("expected all types to be created")
	}

	if len(result.Nodes) != 1 {
		t.Error("expected 1 node")
	}

	if len(result.Ways) != 1 {
		t.Error("expected 1 way")
	}

	if len(result.Relations) != 1 {
		t.Error("expected 1 relation")
	}
}
