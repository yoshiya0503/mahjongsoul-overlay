package ws

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/stretchr/testify/assert"
)

func dialTestServer(t *testing.T, s *httptest.Server) *websocket.Conn {
	t.Helper()
	url := "ws" + strings.TrimPrefix(s.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)
	return conn
}

func newTestServer() (*Hub, *httptest.Server) {
	hub := NewHub()
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		hub.Add(conn)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				hub.Remove(conn)
				conn.Close()
				return
			}
		}
	}))
	return hub, server
}

func TestHub_AddRemove(t *testing.T) {
	hub, server := newTestServer()
	defer server.Close()

	conn := dialTestServer(t, server)
	defer conn.Close()

	// Add されている
	hub.mu.RLock()
	assert.Len(t, hub.clients, 1)
	hub.mu.RUnlock()
}

func TestHub_Broadcast(t *testing.T) {
	hub, server := newTestServer()
	defer server.Close()

	conn1 := dialTestServer(t, server)
	defer conn1.Close()
	conn2 := dialTestServer(t, server)
	defer conn2.Close()

	// 接続が登録されるのを待つ
	time.Sleep(50 * time.Millisecond)

	msg := []byte(`{"test":"hello"}`)
	hub.Broadcast(msg)

	_, data1, err1 := conn1.ReadMessage()
	assert.NoError(t, err1)
	assert.Equal(t, msg, data1)

	_, data2, err2 := conn2.ReadMessage()
	assert.NoError(t, err2)
	assert.Equal(t, msg, data2)
}

func TestHub_BroadcastRemovesDeadConnection(t *testing.T) {
	hub, server := newTestServer()
	defer server.Close()

	conn := dialTestServer(t, server)

	// 接続を閉じる
	conn.Close()

	// サーバー側の ReadMessage ループが閉じを検出して Remove するのを待つ
	time.Sleep(100 * time.Millisecond)

	hub.mu.RLock()
	assert.Empty(t, hub.clients)
	hub.mu.RUnlock()
}
