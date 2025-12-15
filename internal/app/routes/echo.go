package routes

import (
	"log"
	"musick-server/internal/app/services"

	"github.com/DarthPestilane/easytcp"
)

func RegisterEchoRoutes(s *easytcp.Server) {
	s.AddRoute(1, handleEcho)
}

func handleEcho(ctx easytcp.Context) {
	req := ctx.Request()
	log.Printf("received id=%d bytes=%d body=%q", req.ID(), len(req.Data()), string(req.Data()))

	if !services.IsAuthenticated(ctx.Session()) {
		log.Printf("unauthenticated session attempted to use echo route")
		ctx.SetResponseMessage(easytcp.NewMessage(req.ID(), []byte("error: unauthenticated")))
		return
	}
	ctx.SetResponseMessage(easytcp.NewMessage(req.ID(), req.Data()))
}
