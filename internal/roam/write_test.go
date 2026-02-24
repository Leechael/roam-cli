package roam

import "testing"

func TestMoveBlockAction(t *testing.T) {
	a := MoveBlockAction("u1", "p1", "first")
	if a["action"] != "move-block" {
		t.Fatalf("expected move-block action")
	}
	loc := a["location"].(map[string]any)
	if loc["parent-uid"] != "p1" || loc["order"] != "first" {
		t.Fatalf("unexpected location: %#v", loc)
	}
	blk := a["block"].(map[string]any)
	if blk["uid"] != "u1" {
		t.Fatalf("unexpected block uid: %#v", blk)
	}
}
