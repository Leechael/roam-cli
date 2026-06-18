package client

import "testing"

func TestDeletePageAction(t *testing.T) {
	a := DeletePageAction("page-uid")
	if a["action"] != "delete-page" {
		t.Fatalf("expected delete-page action")
	}
	page := a["page"].(map[string]any)
	if page["uid"] != "page-uid" {
		t.Fatalf("unexpected page uid: %#v", page)
	}
}

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
