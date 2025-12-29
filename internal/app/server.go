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
	routes.RegisterRoomRoutes(s)
	routes.RegisterJoinRoomRoutes(s)
	routes.RegisterMessageRoutes(s)
	routes.RegisterShazamRoutes(s)
}
