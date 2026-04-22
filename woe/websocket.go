package woe

import (
	"encoding/json"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// SpinDataWS is command and control
// The server sends this to each client on a spin,
// the spinning client sends this to the server on a spin.
type SpinDataWS struct {
	Type       string    `json:"type"` // "spin" | "sync"
	SpinID     string    `json:"id"`
	SpunString string    `json:"spun"`
	Entries    *[]string `json:"entries"`
	Timestamp  time.Time `json:"timestamp"`
	Velocity   float64   `json:"velocity"`
}

var upgrader = websocket.Upgrader{
	HandshakeTimeout: 5 * time.Second,
	// Allow connections from anywhere
	CheckOrigin: func(r *http.Request) bool { return true },
}

/*

1. Someone clicks spin
2. The websocket receives new spindata
3. The websocket stores state in the wheel struct
4. The websocket sends the new spindata back to all clients
5. JavaScript perceives new data and spins the wheel for everyone
6. Because we're sending the same velocity to all clients, they should all spin to exactly the same spot.

*/

// WSHub is the websocket connection registry
type WSHub struct {
	clients map[*websocket.Conn]bool
	mu      sync.RWMutex
}

// Broadcast writes the message /v/ to all connected clients
func (h *WSHub) Broadcast(v any) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for conn := range h.clients {
		if err := conn.WriteJSON(v); err != nil {
			slog.Error("Could not write websocket message", slog.Any("err", err))
			return
		}
	}
}

func (we *Wheel) WebsocketHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Could not upgrade to websocket", slog.Any("err", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Register ws client on hub
	we.Hub.mu.Lock()
	we.Hub.clients[ws] = true
	we.Hub.mu.Unlock()
	// Deregister on close
	defer func() {
		we.Hub.mu.Lock()
		delete(we.Hub.clients, ws)
		we.Hub.mu.Unlock()
		ws.Close()
	}()

	// Set all joining wheels to this spin data
	// Those that saw a wheel spin should not move
	// And those in after should see what was spun
	// Load Spin data from Wheel
	spindata := SpinDataWS{
		Type:       "sync",
		SpinID:     we.SpinID,
		SpunString: we.SpunString,
		Entries:    we.Entries,
		Timestamp:  time.Now().UTC(),
		Velocity:   we.Velocity,
	}
	if err := ws.WriteJSON(spindata); err != nil {
		slog.Error("Could not sync spin data", slog.Any("err", err))
		ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "Could not send spin data"))
		return
	}

	for {
		msgType, msg, err := ws.ReadMessage()
		if err != nil {
			slog.Error("Could not read websocket message", slog.Any("err", err))
			ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "Could not read websocket message"))
			return
		}
		slog.Info("Received message from websocket", slog.Any("msg", msg))
		switch msgType {
		case websocket.TextMessage:
			var sd SpinDataWS
			if err = json.Unmarshal(msg, &sd); err != nil {
				slog.Error("Could not unmarshal spin data", slog.Any("err", err))
				continue
			}

			if sd.Type == "spin" {
				// New Spin!
				// Coming from the websocket, spindata (sd) remains the same except for velocity
				we.calcVelocity()
				sd.Velocity = we.Velocity

				// Record new state to the wheel
				we.mu.Lock()
				we.SpinID = sd.SpinID
				we.Entries = sd.Entries
				we.SpinTime = sd.Timestamp
				we.mu.Unlock()

				// Init spin to clients
				we.Hub.Broadcast(sd)
			}
		}
	}
}

// calcVelocity writes a new 'fuzzy' velocity to the wheel
func (we *Wheel) calcVelocity() {
	velocity := 0.05
	rnd, err := RandVelocity(0.20, 0.30)
	if err != nil {
		rnd = 0.42
		slog.Error("Could not generate random velocity, using default", slog.Any("err", err), slog.Float64("default", rnd))
	}
	velocity += rnd

	we.mu.Lock()
	we.Velocity += velocity
	we.mu.Unlock()
}

// RandVelocity returns a random velocity between min and max
func RandVelocity(min, max float64) (float64, error) {
	rnd := min + rand.Float64()*(max-min)
	r := strconv.FormatFloat(rnd, 'f', 2, 64)
	re, err := strconv.ParseFloat(r, 64)
	if err != nil {
		slog.Error("Could not parse velocity", slog.Any("err", err))
		return 0, err
	}
	slog.Debug("+++ Random velocity +++", slog.Any("velocity", re))
	return re, nil
}
