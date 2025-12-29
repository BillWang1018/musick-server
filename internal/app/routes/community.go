package routes

import (
	"encoding/json"
	"log"
	"time"

	"musick-server/internal/app/services"

	"github.com/DarthPestilane/easytcp"
)

// Community post DTOs

type CreatePostRequest struct {
	UserID string `json:"user_id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

type CreatePostResponse struct {
	Success bool                    `json:"success"`
	Message string                  `json:"message"`
	Post    *services.CommunityPost `json:"post,omitempty"`
}

type DeletePostRequest struct {
	UserID string `json:"user_id"`
	PostID string `json:"post_id"`
}

type DeletePostResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type UpdatePostRequest struct {
	UserID string  `json:"user_id"`
	PostID string  `json:"post_id"`
	Title  *string `json:"title,omitempty"`
	Body   *string `json:"body,omitempty"`
}

type UpdatePostResponse struct {
	Success bool                    `json:"success"`
	Message string                  `json:"message"`
	Post    *services.CommunityPost `json:"post,omitempty"`
}

type ListPostsRequest struct {
	UserID            string `json:"user_id"`
	BeforeID          string `json:"before_id"`
	Limit             int    `json:"limit"`
	IncludeAttachment bool   `json:"include_attachment"`
}

type ListPostsResponse struct {
	Success    bool       `json:"success"`
	Message    string     `json:"message"`
	Posts      []PostItem `json:"posts,omitempty"`
	HasMore    bool       `json:"has_more"`
	NextBefore string     `json:"next_before,omitempty"`
}

type PostItem struct {
	ID          string                         `json:"id"`
	AuthorID    string                         `json:"author_id"`
	AuthorName  string                         `json:"author_name,omitempty"`
	Title       string                         `json:"title"`
	Body        string                         `json:"body"`
	CreatedAt   string                         `json:"created_at"`
	UpdatedAt   string                         `json:"updated_at"`
	Attachments []services.CommunityAttachment `json:"attachments,omitempty"`
}

// RegisterCommunityRoutes wires post-related handlers.
func RegisterCommunityRoutes(s *easytcp.Server) {
	s.AddRoute(701, handleCreatePost)
	s.AddRoute(702, handleDeletePost)
	s.AddRoute(710, handleListPosts)
	s.AddRoute(711, handleUpdatePost)
}

func handleCreatePost(ctx easytcp.Context) {
	req := ctx.Request()
	log.Printf("701 create post: id=%d bytes=%d", req.ID(), len(req.Data()))

	if !services.IsAuthenticated(ctx.Session()) {
		sendCreatePostError(ctx, "not authenticated")
		return
	}

	var createReq CreatePostRequest
	if err := json.Unmarshal(req.Data(), &createReq); err != nil {
		sendCreatePostError(ctx, "invalid request format")
		return
	}

	if createReq.UserID == "" || createReq.Title == "" || createReq.Body == "" {
		sendCreatePostError(ctx, "user_id, title, and body are required")
		return
	}

	session := services.GetSession(ctx.Session())
	if session == nil || session.UserID != createReq.UserID {
		sendCreatePostError(ctx, "user_id mismatch")
		return
	}

	post, err := services.CreateCommunityPost(createReq.UserID, createReq.Title, createReq.Body)
	if err != nil {
		log.Printf("failed to create post: %v", err)
		sendCreatePostError(ctx, "failed to create post")
		return
	}

	resp := CreatePostResponse{Success: true, Message: "post created", Post: post}
	data, _ := json.Marshal(resp)
	ctx.SetResponseMessage(easytcp.NewMessage(req.ID(), data))
}

func sendCreatePostError(ctx easytcp.Context, msg string) {
	resp := CreatePostResponse{Success: false, Message: msg}
	data, _ := json.Marshal(resp)
	ctx.SetResponseMessage(easytcp.NewMessage(ctx.Request().ID(), data))
}

func handleDeletePost(ctx easytcp.Context) {
	req := ctx.Request()
	log.Printf("702 delete post: id=%d bytes=%d", req.ID(), len(req.Data()))

	if !services.IsAuthenticated(ctx.Session()) {
		sendDeletePostError(ctx, "not authenticated")
		return
	}

	var delReq DeletePostRequest
	if err := json.Unmarshal(req.Data(), &delReq); err != nil {
		sendDeletePostError(ctx, "invalid request format")
		return
	}

	if delReq.UserID == "" || delReq.PostID == "" {
		sendDeletePostError(ctx, "user_id and post_id are required")
		return
	}

	session := services.GetSession(ctx.Session())
	if session == nil || session.UserID != delReq.UserID {
		sendDeletePostError(ctx, "user_id mismatch")
		return
	}

	if err := services.DeleteCommunityPost(delReq.PostID, delReq.UserID); err != nil {
		log.Printf("failed to delete post: %v", err)
		sendDeletePostError(ctx, "failed to delete post")
		return
	}

	resp := DeletePostResponse{Success: true, Message: "post deleted"}
	data, _ := json.Marshal(resp)
	ctx.SetResponseMessage(easytcp.NewMessage(req.ID(), data))
}

func sendDeletePostError(ctx easytcp.Context, msg string) {
	resp := DeletePostResponse{Success: false, Message: msg}
	data, _ := json.Marshal(resp)
	ctx.SetResponseMessage(easytcp.NewMessage(ctx.Request().ID(), data))
}

func handleUpdatePost(ctx easytcp.Context) {
	req := ctx.Request()
	log.Printf("711 update post: id=%d bytes=%d", req.ID(), len(req.Data()))

	if !services.IsAuthenticated(ctx.Session()) {
		sendUpdatePostError(ctx, "not authenticated")
		return
	}

	var updReq UpdatePostRequest
	if err := json.Unmarshal(req.Data(), &updReq); err != nil {
		sendUpdatePostError(ctx, "invalid request format")
		return
	}

	if updReq.UserID == "" || updReq.PostID == "" {
		sendUpdatePostError(ctx, "user_id and post_id are required")
		return
	}

	if (updReq.Title == nil || *updReq.Title == "") && (updReq.Body == nil || *updReq.Body == "") {
		sendUpdatePostError(ctx, "title or body must be provided")
		return
	}

	session := services.GetSession(ctx.Session())
	if session == nil || session.UserID != updReq.UserID {
		sendUpdatePostError(ctx, "user_id mismatch")
		return
	}

	post, err := services.UpdateCommunityPost(updReq.PostID, updReq.UserID, updReq.Title, updReq.Body)
	if err != nil {
		log.Printf("failed to update post: %v", err)
		sendUpdatePostError(ctx, "failed to update post")
		return
	}

	resp := UpdatePostResponse{Success: true, Message: "post updated", Post: post}
	data, _ := json.Marshal(resp)
	ctx.SetResponseMessage(easytcp.NewMessage(req.ID(), data))
}

func sendUpdatePostError(ctx easytcp.Context, msg string) {
	resp := UpdatePostResponse{Success: false, Message: msg}
	data, _ := json.Marshal(resp)
	ctx.SetResponseMessage(easytcp.NewMessage(ctx.Request().ID(), data))
}

func handleListPosts(ctx easytcp.Context) {
	req := ctx.Request()
	log.Printf("710 list posts: id=%d bytes=%d", req.ID(), len(req.Data()))

	if !services.IsAuthenticated(ctx.Session()) {
		sendListPostsError(ctx, "not authenticated")
		return
	}

	var lpReq ListPostsRequest
	if err := json.Unmarshal(req.Data(), &lpReq); err != nil {
		sendListPostsError(ctx, "invalid request format")
		return
	}

	if lpReq.UserID == "" {
		sendListPostsError(ctx, "user_id is required")
		return
	}

	session := services.GetSession(ctx.Session())
	if session == nil || session.UserID != lpReq.UserID {
		sendListPostsError(ctx, "user_id mismatch")
		return
	}

	posts, hasMore, err := services.ListCommunityPosts(lpReq.BeforeID, lpReq.Limit, lpReq.IncludeAttachment)
	if err != nil {
		log.Printf("failed to list posts: %v", err)
		sendListPostsError(ctx, "failed to list posts")
		return
	}

	items := make([]PostItem, 0, len(posts))
	for _, p := range posts {
		items = append(items, PostItem{
			ID:          p.ID,
			AuthorID:    p.AuthorID,
			AuthorName:  p.AuthorName,
			Title:       p.Title,
			Body:        p.Body,
			CreatedAt:   p.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   p.UpdatedAt.Format(time.RFC3339),
			Attachments: p.Attachments,
		})
	}

	resp := ListPostsResponse{
		Success: true,
		Message: "posts fetched",
		Posts:   items,
		HasMore: hasMore,
	}

	if len(posts) > 0 {
		last := posts[len(posts)-1]
		resp.NextBefore = last.CreatedAt.Format(time.RFC3339)
	}

	data, _ := json.Marshal(resp)
	ctx.SetResponseMessage(easytcp.NewMessage(req.ID(), data))
}

func sendListPostsError(ctx easytcp.Context, msg string) {
	resp := ListPostsResponse{Success: false, Message: msg}
	data, _ := json.Marshal(resp)
	ctx.SetResponseMessage(easytcp.NewMessage(ctx.Request().ID(), data))
}
