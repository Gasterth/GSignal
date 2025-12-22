package gsignal_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Merluz/gsignal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	TestEventA = "TEST_EVENT_A"
	TestEventB = "TEST_EVENT_B"
)

// TestHub_SubscribeAndPublish_Sync verifies that a synchronous publish
// is correctly received by a subscriber.
func TestHub_SubscribeAndPublish_Sync(t *testing.T) {
	hub := gsignal.New()
	defer hub.Shutdown()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub, err := hub.Subscribe(ctx, TestEventA)
	require.NoError(t, err)
	require.NotNil(t, sub)

	evt, err := gsignal.NewEvent(TestEventA, "hello world")
	require.NoError(t, err)

	err = hub.Publish(evt)
	assert.NoError(t, err)

	select {
	case received := <-sub.Events():
		assert.Equal(t, evt.ID, received.ID)
		assert.Equal(t, TestEventA, received.Type)
		assert.Equal(t, "hello world", received.Payload)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for event")
	}
}

// TestHub_GlobalSubscription verifies that a global subscription (no types)
// receives all events published to the hub.
func TestHub_GlobalSubscription(t *testing.T) {
	hub := gsignal.New()
	defer hub.Shutdown()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub, err := hub.Subscribe(ctx)
	require.NoError(t, err)

	evtA, _ := gsignal.NewEvent(TestEventA, "A")
	evtB, _ := gsignal.NewEvent(TestEventB, "B")

	_ = hub.Publish(evtA)
	_ = hub.Publish(evtB)

	received := 0
	for i := 0; i < 2; i++ {
		select {
		case <-sub.Events():
			received++
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for events")
		}
	}

	assert.Equal(t, 2, received)
}

// TestHub_PublishAsync verifies that asynchronous publishing works
// and eventually delivers the event to the subscriber.
func TestHub_PublishAsync(t *testing.T) {
	hub := gsignal.New()
	defer hub.Shutdown()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub, err := hub.Subscribe(ctx, TestEventA)
	require.NoError(t, err)

	evt, _ := gsignal.NewEvent(TestEventA, "async payload")
	err = hub.PublishAsync(evt)
	assert.NoError(t, err)

	select {
	case received := <-sub.Events():
		assert.Equal(t, "async payload", received.Payload)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for async event")
	}
}

// TestHub_Unsubscribe verifies that removing a subscription stops
// the delivery of further events and closes the channel.
func TestHub_Unsubscribe(t *testing.T) {
	hub := gsignal.New()
	defer hub.Shutdown()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub, err := hub.Subscribe(ctx, TestEventA)
	require.NoError(t, err)

	err = hub.Unsubscribe(sub)
	assert.NoError(t, err)

	select {
	case _, ok := <-sub.Events():
		assert.False(t, ok, "subscription channel should be closed")
	default:
	}
}

// TestHub_ContextCancellation verifies that cancelling the context
// automatically closes the associated subscription.
func TestHub_ContextCancellation(t *testing.T) {
	hub := gsignal.New()
	defer hub.Shutdown()

	ctx, cancel := context.WithCancel(context.Background())
	sub, err := hub.Subscribe(ctx, TestEventA)
	require.NoError(t, err)

	cancel()
	time.Sleep(50 * time.Millisecond)

	assert.True(t, sub.IsClosed())
}

// TestHub_Shutdown verifies that shutting down the hub:
// 1. Closes all subscriptions
// 2. Prevents new publications
func TestHub_Shutdown(t *testing.T) {
	hub := gsignal.New()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub, _ := hub.Subscribe(ctx, TestEventA)

	err := hub.Shutdown()
	assert.NoError(t, err)

	assert.True(t, hub.Closed())
	assert.True(t, sub.IsClosed())

	evt, _ := gsignal.NewEvent(TestEventA, "fail")
	err = hub.Publish(evt)
	assert.ErrorIs(t, err, gsignal.ErrHubClosed)
}

// TestHub_Concurrency verifies that the Hub is thread-safe
// under heavy concurrent load (multiple subscribers and publishers).
func TestHub_Concurrency(t *testing.T) {
	hub := gsignal.New()
	defer hub.Shutdown()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const numSubs = 10
	const numEvents = 50

	subs := make([]gsignal.Subscription, numSubs)
	for i := 0; i < numSubs; i++ {
		sub, err := hub.Subscribe(ctx, TestEventA)
		require.NoError(t, err)
		subs[i] = sub
	}

	var wg sync.WaitGroup
	for _, s := range subs {
		wg.Add(1)
		go func(sub gsignal.Subscription) {
			defer wg.Done()
			count := 0
			for range sub.Events() {
				count++
				if count == numEvents {
					return
				}
			}
		}(s)
	}

	time.Sleep(10 * time.Millisecond)

	for i := 0; i < numEvents; i++ {
		evt, _ := gsignal.NewEvent(TestEventA, i)
		_ = hub.Publish(evt)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("concurrency test timed out")
	}
}
