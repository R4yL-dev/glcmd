package events

import (
	"log/slog"
	"sync"
	"testing"
	"time"
)

func TestBroker_SubscribeUnsubscribe(t *testing.T) {
	broker := NewBroker(10, slog.Default())

	// Subscribe
	ch := broker.Subscribe("client1", nil)
	if ch == nil {
		t.Fatal("expected non-nil channel")
	}
	if broker.SubscriberCount() != 1 {
		t.Errorf("expected 1 subscriber, got %d", broker.SubscriberCount())
	}

	// Subscribe another
	ch2 := broker.Subscribe("client2", []EventType{EventTypeGlucose})
	if ch2 == nil {
		t.Fatal("expected non-nil channel")
	}
	if broker.SubscriberCount() != 2 {
		t.Errorf("expected 2 subscribers, got %d", broker.SubscriberCount())
	}

	// Unsubscribe
	broker.Unsubscribe("client1")
	if broker.SubscriberCount() != 1 {
		t.Errorf("expected 1 subscriber after unsubscribe, got %d", broker.SubscriberCount())
	}

	// Verify channel is closed
	_, ok := <-ch
	if ok {
		t.Error("expected channel to be closed")
	}

	broker.Unsubscribe("client2")
	if broker.SubscriberCount() != 0 {
		t.Errorf("expected 0 subscribers, got %d", broker.SubscriberCount())
	}
}

func TestBroker_PublishToAllSubscribers(t *testing.T) {
	broker := NewBroker(10, slog.Default())

	ch1 := broker.Subscribe("client1", nil)
	ch2 := broker.Subscribe("client2", nil)

	event := Event{Type: EventTypeGlucose, Data: "test"}
	broker.Publish(event)

	// Both should receive
	select {
	case e := <-ch1:
		if e.Type != EventTypeGlucose {
			t.Errorf("expected glucose event, got %s", e.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("client1 did not receive event")
	}

	select {
	case e := <-ch2:
		if e.Type != EventTypeGlucose {
			t.Errorf("expected glucose event, got %s", e.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("client2 did not receive event")
	}

	broker.Unsubscribe("client1")
	broker.Unsubscribe("client2")
}

func TestBroker_PublishWithTypeFilter(t *testing.T) {
	broker := NewBroker(10, slog.Default())

	// client1 wants only glucose
	ch1 := broker.Subscribe("client1", []EventType{EventTypeGlucose})
	// client2 wants only sensor
	ch2 := broker.Subscribe("client2", []EventType{EventTypeSensor})
	// client3 wants all
	ch3 := broker.Subscribe("client3", nil)

	// Publish glucose event
	broker.Publish(Event{Type: EventTypeGlucose, Data: "glucose"})

	// client1 should receive
	select {
	case e := <-ch1:
		if e.Type != EventTypeGlucose {
			t.Errorf("client1: expected glucose, got %s", e.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("client1 did not receive glucose event")
	}

	// client2 should NOT receive
	select {
	case <-ch2:
		t.Error("client2 should not receive glucose event")
	case <-time.After(50 * time.Millisecond):
		// Expected
	}

	// client3 should receive
	select {
	case e := <-ch3:
		if e.Type != EventTypeGlucose {
			t.Errorf("client3: expected glucose, got %s", e.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("client3 did not receive glucose event")
	}

	// Publish sensor event
	broker.Publish(Event{Type: EventTypeSensor, Data: "sensor"})

	// client1 should NOT receive
	select {
	case <-ch1:
		t.Error("client1 should not receive sensor event")
	case <-time.After(50 * time.Millisecond):
		// Expected
	}

	// client2 should receive
	select {
	case e := <-ch2:
		if e.Type != EventTypeSensor {
			t.Errorf("client2: expected sensor, got %s", e.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("client2 did not receive sensor event")
	}

	// client3 should receive
	select {
	case e := <-ch3:
		if e.Type != EventTypeSensor {
			t.Errorf("client3: expected sensor, got %s", e.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("client3 did not receive sensor event")
	}

	broker.Unsubscribe("client1")
	broker.Unsubscribe("client2")
	broker.Unsubscribe("client3")
}

func TestBroker_NonBlockingPublish(t *testing.T) {
	// Create broker with small buffer
	broker := NewBroker(2, slog.Default())

	ch := broker.Subscribe("slow-client", nil)

	// Fill the buffer
	broker.Publish(Event{Type: EventTypeGlucose, Data: "1"})
	broker.Publish(Event{Type: EventTypeGlucose, Data: "2"})

	// This should not block even though buffer is full
	done := make(chan struct{})
	go func() {
		broker.Publish(Event{Type: EventTypeGlucose, Data: "3"})
		close(done)
	}()

	select {
	case <-done:
		// Expected - publish returned without blocking
	case <-time.After(100 * time.Millisecond):
		t.Error("Publish blocked on full buffer")
	}

	// Drain channel
	<-ch
	<-ch

	broker.Unsubscribe("slow-client")
}

func TestBroker_ConcurrentAccess(t *testing.T) {
	broker := NewBroker(100, slog.Default())

	var wg sync.WaitGroup
	const numClients = 10
	const numEvents = 100

	// Start multiple subscribers
	channels := make([]<-chan Event, numClients)
	for i := 0; i < numClients; i++ {
		id := string(rune('a' + i))
		channels[i] = broker.Subscribe(id, nil)
	}

	// Publish events concurrently
	for i := 0; i < numEvents; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			broker.Publish(Event{Type: EventTypeGlucose, Data: n})
		}(i)
	}

	wg.Wait()

	// Unsubscribe all
	for i := 0; i < numClients; i++ {
		id := string(rune('a' + i))
		broker.Unsubscribe(id)
	}

	if broker.SubscriberCount() != 0 {
		t.Errorf("expected 0 subscribers, got %d", broker.SubscriberCount())
	}
}

func TestBroker_StartStop(t *testing.T) {
	broker := NewBroker(10, slog.Default())

	ch := broker.Subscribe("client", nil)
	broker.Start()

	// Wait a bit less than heartbeat interval, should not receive keepalive yet
	select {
	case <-ch:
		t.Error("should not receive event before heartbeat interval")
	case <-time.After(50 * time.Millisecond):
		// Expected
	}

	// Stop should close channels
	broker.Stop()

	// Channel should be closed
	_, ok := <-ch
	if ok {
		t.Error("expected channel to be closed after Stop")
	}
}

func TestBroker_MultipleTypeFilter(t *testing.T) {
	broker := NewBroker(10, slog.Default())

	// Subscribe to both glucose and sensor
	ch := broker.Subscribe("client", []EventType{EventTypeGlucose, EventTypeSensor})

	// Should receive glucose
	broker.Publish(Event{Type: EventTypeGlucose, Data: "glucose"})
	select {
	case e := <-ch:
		if e.Type != EventTypeGlucose {
			t.Errorf("expected glucose, got %s", e.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("did not receive glucose event")
	}

	// Should receive sensor
	broker.Publish(Event{Type: EventTypeSensor, Data: "sensor"})
	select {
	case e := <-ch:
		if e.Type != EventTypeSensor {
			t.Errorf("expected sensor, got %s", e.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("did not receive sensor event")
	}

	// Should NOT receive keepalive
	broker.Publish(Event{Type: EventTypeKeepalive, Data: nil})
	select {
	case <-ch:
		t.Error("should not receive keepalive event")
	case <-time.After(50 * time.Millisecond):
		// Expected
	}

	broker.Unsubscribe("client")
}

func TestSubscriber_WantsEvent(t *testing.T) {
	tests := []struct {
		name      string
		types     []EventType
		eventType EventType
		want      bool
	}{
		{"empty filter accepts all", nil, EventTypeGlucose, true},
		{"empty filter accepts keepalive", nil, EventTypeKeepalive, true},
		{"glucose filter accepts glucose", []EventType{EventTypeGlucose}, EventTypeGlucose, true},
		{"glucose filter rejects sensor", []EventType{EventTypeGlucose}, EventTypeSensor, false},
		{"multi filter accepts glucose", []EventType{EventTypeGlucose, EventTypeSensor}, EventTypeGlucose, true},
		{"multi filter accepts sensor", []EventType{EventTypeGlucose, EventTypeSensor}, EventTypeSensor, true},
		{"multi filter rejects keepalive", []EventType{EventTypeGlucose, EventTypeSensor}, EventTypeKeepalive, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := &Subscriber{Types: tt.types}
			got := sub.wantsEvent(tt.eventType)
			if got != tt.want {
				t.Errorf("wantsEvent(%s) = %v, want %v", tt.eventType, got, tt.want)
			}
		})
	}
}
