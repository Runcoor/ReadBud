// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package sse

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

const redisSSEPrefix = "sse:"

// redisMessage is the wire format for Redis pub/sub SSE events.
type redisMessage struct {
	Channel string `json:"channel"`
	Event   Event  `json:"event"`
}

// RedisPublisher publishes SSE events via Redis pub/sub.
// Used by the worker process which has no direct SSE clients.
type RedisPublisher struct {
	rdb *redis.Client
}

// NewRedisPublisher creates a new Redis-backed event publisher.
func NewRedisPublisher(rdb *redis.Client) *RedisPublisher {
	return &RedisPublisher{rdb: rdb}
}

// Publish sends an event to Redis pub/sub so the API process can forward it.
func (p *RedisPublisher) Publish(channel string, event Event) {
	msg := redisMessage{Channel: channel, Event: event}
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	p.rdb.Publish(context.Background(), redisSSEPrefix+"events", data)
}

// StartRedisSubscriber subscribes to Redis pub/sub and forwards events to the local Hub.
// This runs in the API process. Call in a goroutine.
func StartRedisSubscriber(ctx context.Context, rdb *redis.Client, hub *Hub) {
	sub := rdb.Subscribe(ctx, redisSSEPrefix+"events")
	defer sub.Close()

	ch := sub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			var rm redisMessage
			if err := json.Unmarshal([]byte(msg.Payload), &rm); err != nil {
				log.Printf("sse: redis unmarshal error: %v", err)
				continue
			}
			hub.Publish(rm.Channel, rm.Event)
		}
	}
}

// NewRedisClient creates a Redis client from config values.
func NewRedisClient(addr, password string, db int) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
}

// Compile-time check.
var _ EventPublisher = (*RedisPublisher)(nil)
var _ EventPublisher = (*Hub)(nil)
var _ fmt.Stringer = (*redisMessage)(nil) // suppress unused import

func (m redisMessage) String() string { return m.Channel }
