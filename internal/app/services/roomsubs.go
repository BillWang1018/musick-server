package services

import (
	"log"
	"sync"

	"github.com/DarthPestilane/easytcp"
)

// roomSubs tracks active sessions per room for broadcasting.
var (
	roomSubs   = make(map[string]map[interface{}]easytcp.Session)
	roomSubsMu sync.RWMutex
	roomPacker = easytcp.NewDefaultPacker()
)

// AddSessionToRoom tracks a session as present in a room.
func AddSessionToRoom(roomID string, sess easytcp.Session) {
	roomSubsMu.Lock()
	defer roomSubsMu.Unlock()
	if roomSubs[roomID] == nil {
		roomSubs[roomID] = make(map[interface{}]easytcp.Session)
	}
	roomSubs[roomID][sess.ID()] = sess
}

// RemoveSessionFromRoom removes a session from a specific room.
func RemoveSessionFromRoom(roomID string, sess easytcp.Session) {
	roomSubsMu.Lock()
	defer roomSubsMu.Unlock()
	if subs, ok := roomSubs[roomID]; ok {
		delete(subs, sess.ID())
		if len(subs) == 0 {
			delete(roomSubs, roomID)
		}
	}
}

// RemoveSessionFromAllRooms removes a session from all tracked rooms (on disconnect).
func RemoveSessionFromAllRooms(sess easytcp.Session) {
	roomSubsMu.Lock()
	defer roomSubsMu.Unlock()
	for roomID, subs := range roomSubs {
		delete(subs, sess.ID())
		if len(subs) == 0 {
			delete(roomSubs, roomID)
		}
	}
}

// BroadcastToRoom sends a message to all sessions tracked in the room.
// If skipID is non-nil, that session ID will not receive the broadcast.
func BroadcastToRoom(roomID string, msg *easytcp.Message, skipID interface{}) {
	roomSubsMu.RLock()
	subs := roomSubs[roomID]
	roomSubsMu.RUnlock()
	if len(subs) == 0 {
		return
	}

	data, err := roomPacker.Pack(msg)
	if err != nil {
		log.Printf("broadcast pack failed for room %s: %v", roomID, err)
		return
	}

	for id, sess := range subs {
		if skipID != nil && id == skipID {
			continue
		}
		if _, err := sess.Conn().Write(data); err != nil {
			log.Printf("broadcast to room %s failed for session %v: %v", roomID, id, err)
		}
	}
}
