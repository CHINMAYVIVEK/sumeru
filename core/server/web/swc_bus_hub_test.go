package web

import (
	"encoding/json"
	"testing"
)

func TestSwcBusHubBroadcastFiltersByActor(t *testing.T) {
	h := &busHub{clients: make(map[*swcBusClient]struct{})}
	userA := &swcBusClient{uid: 1, send: make(chan []byte, 1)}
	userB := &swcBusClient{uid: 2, send: make(chan []byte, 1)}
	h.register(userA)
	h.register(userB)

	msg := []byte(`{"channel":"record.updated","payload":{"model":"crm.lead","id":1}}`)
	h.broadcast(1, msg)

	select {
	case got := <-userA.send:
		if string(got) != string(msg) {
			t.Fatalf("user A: got %q", got)
		}
	default:
		t.Fatal("expected message for user A")
	}
	select {
	case <-userB.send:
		t.Fatal("user B should not receive actor-scoped message")
	default:
	}
}

func TestSwcBusHubQueueMessageShape(t *testing.T) {
	h := &busHub{clients: make(map[*swcBusClient]struct{})}
	client := &swcBusClient{uid: 5, send: make(chan []byte, 1)}
	h.register(client)

	raw, _ := json.Marshal(map[string]interface{}{
		"name":    "record.updated",
		"payload": map[string]interface{}{"model": "core.partner", "id": 3},
		"actor":   5,
	})
	var envelope map[string]interface{}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatal(err)
	}
	name := envelope["name"].(string)
	inner := envelope["payload"].(map[string]interface{})
	out, _ := json.Marshal(map[string]interface{}{"channel": name, "payload": inner})
	h.broadcast(5, out)

	select {
	case got := <-client.send:
		var parsed map[string]interface{}
		if err := json.Unmarshal(got, &parsed); err != nil {
			t.Fatal(err)
		}
		if parsed["channel"] != "record.updated" {
			t.Fatalf("channel: %v", parsed["channel"])
		}
	default:
		t.Fatal("expected bus message")
	}
}
