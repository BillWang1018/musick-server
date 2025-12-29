package app

import (
	"log"

	"musick-server/internal/app/routes"
	"musick-server/internal/app/services"

	"github.com/DarthPestilane/easytcp"
)

// Server wraps easytcp.Server and centralizes route registration.
type Server struct {
	srv *easytcp.Server
}

// New creates a configured server and registers all routes.
func New() *Server {
	// 1. 建立 DefaultPacker 實例
	packer := easytcp.NewDefaultPacker()

	// 2. 關鍵修正：將最大封包限制調大至 10MB (預設可能太小導致斷線)
	packer.MaxDataSize = 10 * 1024 * 1024

	// 3. 將設定好的 packer 傳入 ServerOption
	srv := easytcp.NewServer(&easytcp.ServerOption{
		Packer: packer,
	})

	// Log when clients connect/disconnect.
	srv.OnSessionCreate = func(sess easytcp.Session) {
		addr := sess.Conn().RemoteAddr().String()
		log.Printf("client connected: %s", addr)
	}
	srv.OnSessionClose = func(sess easytcp.Session) {
		addr := sess.Conn().RemoteAddr().String()
		log.Printf("client disconnected: %s", addr)
		services.RemoveSession(sess)
		services.RemoveSessionFromAllRooms(sess)
	}

	registerRoutes(srv)

	return &Server{srv: srv}
}

// Run starts listening on the provided address.
func (s *Server) Run(addr string) error {
	log.Printf("listening on %s", addr)
	return s.srv.Run(addr)
}

// registerRoutes wires all message handlers.
func registerRoutes(s *easytcp.Server) {
	routes.RegisterEchoRoutes(s)
	routes.RegisterAuthRoutes(s)

	// Route 201: create room.
	// Route 210: list rooms.
	// Route 211: find public rooms.
	routes.RegisterRoomRoutes(s)
	routes.RegisterJoinRoomRoutes(s)
	routes.RegisterMessageRoutes(s)

	// Route 501: create song; 510: list songs; 511: update song.
	routes.RegisterSongRoutes(s)

	// Route 601: create note; 602: delete note; 603: broadcast note updates; 610: list notes.
	routes.RegisterNoteRoutes(s)

	// Route 604: create track; 605: delete track; 606: broadcast track updates.
	routes.RegisterTrackRoutes(s)

	// Route 701: create post; 702: delete post; 710: list posts; 711: update post.
	routes.RegisterCommunityRoutes(s)
	routes.RegisterShazamRoutes(s)
}
