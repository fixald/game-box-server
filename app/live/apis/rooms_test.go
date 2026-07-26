package apis

import (
	"encoding/json"
	"testing"
)

func TestPrefixedID(t *testing.T) {
	if got := prefixedID("game", nil); got != nil {
		t.Fatalf("nil ID should remain nil, got %v", *got)
	}
	id := uint(1001)
	got := prefixedID("game", &id)
	if got == nil || *got != "game_1001" {
		t.Fatalf("prefixed ID = %v, want game_1001", got)
	}
}

func TestRoomItemJSONContract(t *testing.T) {
	id := "game_1001"
	payload, err := json.Marshal(RoomItem{ID: "live_1001", GameID: &id, Sort: 1})
	if err != nil {
		t.Fatal(err)
	}
	if string(payload) == "" || !contains(string(payload), `"gameId":"game_1001"`) || !contains(string(payload), `"sort":1`) {
		t.Fatalf("unexpected room JSON: %s", payload)
	}
}

func contains(value, part string) bool {
	for i := 0; i+len(part) <= len(value); i++ {
		if value[i:i+len(part)] == part {
			return true
		}
	}
	return false
}
