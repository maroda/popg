package woe_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	wheel "github.com/maroda/popg/woe"
)

func TestWebsocketHandler(t *testing.T) {
	// Create and configure a wheel
	we, err := wheel.NewWheel(&[]string{"one, two"})
	assertError(t, err, nil)

	// Create a websocket server
	serve := httptest.NewServer(http.HandlerFunc(we.WebsocketHandler))
	defer serve.Close()

	// Connect to handler function as a WS client
	url := "ws" + strings.TrimPrefix(serve.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	assertError(t, err, nil)
	defer ws.Close()

	// Write a message
	// This simulates the spin button being pushed
	// So it's all new data
	words := []string{"one", "two", "three"}
	spindata := &wheel.SpinDataWS{
		Type:      "spin",
		SpinID:    uuid.New().String(), // New spin, so we need to send a new ID
		Entries:   &words,              // Contents of the wheel
		Timestamp: time.Now().UTC(),
	}
	err = ws.WriteJSON(spindata)
	assertError(t, err, nil)

	result := &wheel.SpinDataWS{}
	err = ws.ReadJSON(result)
	assertError(t, err, nil)
	t.Log(result)

	err = ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	assertError(t, err, nil)
}

func TestWheel_RandVelocity(t *testing.T) {
	var got []float64

	count := 20

	t.Log("Getting random numbers...")
	for i := 0; i < count; i++ {
		r, err := wheel.RandVelocity(0.25, 0.35)
		assertError(t, err, nil)
		got = append(got, r)
		time.Sleep(100 * time.Millisecond)
	}
	assertInt(t, count, len(got))

	// state is for comparison,
	// variability scores how many "in a row" dupes happened
	state, variability := 0.0, 0
	for _, v := range got {
		if state == 0.0 {
			state = v
		} else {
			if v == state {
				variability--
			} else if v != state {
				variability++
			}
			state = v
		}
	}

	if variability == 0 {
		t.Errorf("variability is zero: %d", variability)
	} else if variability < (count/2)-1 {
		t.Errorf("variability is low: %d", variability)
	}
}
