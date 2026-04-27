// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package sse

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

// Event represents a server-sent event.
type Event struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// EventPublisher is the interface for publishing SSE events.
// Both Hub (in-process) and RedisPublisher (cross-process) implement this.
type EventPublisher interface {
	Publish(channel string, event Event)
}

// Hub manages SSE connections for multiple task channels.
type Hub struct {
	mu       sync.RWMutex
	channels map[string]map[chan Event]struct{}
}

// NewHub creates a new SSE hub.
func NewHub() *Hub {
	return &Hub{
		channels: make(map[string]map[chan Event]struct{}),
	}
}

// Subscribe registers a new listener for a channel (e.g., task ID).
func (h *Hub) Subscribe(channel string) chan Event {
	h.mu.Lock()
	defer h.mu.Unlock()

	ch := make(chan Event, 16)
	if h.channels[channel] == nil {
		h.channels[channel] = make(map[chan Event]struct{})
	}
	h.channels[channel][ch] = struct{}{}
	return ch
}

// Unsubscribe removes a listener from a channel.
func (h *Hub) Unsubscribe(channel string, ch chan Event) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if subs, ok := h.channels[channel]; ok {
		delete(subs, ch)
		close(ch)
		if len(subs) == 0 {
			delete(h.channels, channel)
		}
	}
}

// Publish sends an event to all listeners on a channel.
func (h *Hub) Publish(channel string, event Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if subs, ok := h.channels[channel]; ok {
		for ch := range subs {
			select {
			case ch <- event:
			default:
				// Skip slow consumers
			}
		}
	}
}

// ServeHTTP creates a Gin handler for SSE streaming on a given channel key param.
func (h *Hub) ServeHTTP(paramName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		channel := c.Param(paramName)
		if channel == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing channel"})
			return
		}

		// Set SSE headers
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("X-Accel-Buffering", "no")

		ch := h.Subscribe(channel)
		defer h.Unsubscribe(channel, ch)

		c.Stream(func(w io.Writer) bool {
			select {
			case event, ok := <-ch:
				if !ok {
					return false
				}
				data, _ := json.Marshal(event)
				fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, string(data))
				return true
			case <-c.Request.Context().Done():
				return false
			}
		})
	}
}
