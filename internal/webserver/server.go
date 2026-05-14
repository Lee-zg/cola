// 文件说明：internal/webserver/server.go，负责应用后端或核心业务实现。
package webserver

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"cola/internal/bookmark"
	"cola/internal/exporter"
)

type CatalogStore interface {
	ListBookmarks(ctx context.Context, req bookmark.SearchRequest) (bookmark.SearchResult, error)
	AllBookmarks(ctx context.Context) ([]bookmark.Bookmark, error)
	ListFolders(ctx context.Context) ([]string, error)
	ListTags(ctx context.Context) ([]string, error)
}

type Server struct {
	store    CatalogStore
	mu       sync.Mutex
	server   *http.Server
	listener net.Listener
	url      string
}

func New(store CatalogStore) *Server {
	return &Server{store: store}
}

func (s *Server) Start(ctx context.Context) (bookmark.ServerStatus, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.server != nil {
		return s.statusLocked(), nil
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return bookmark.ServerStatus{}, err
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/api/bookmarks", s.handleBookmarks)
	mux.HandleFunc("/api/meta", s.handleMeta)
	s.server = &http.Server{
		Handler:           securityHeaders(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}
	s.listener = listener
	s.url = "http://" + listener.Addr().String()
	go func() {
		if err := s.server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			println("local web server error:", err.Error())
		}
	}()
	return s.statusLocked(), nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.mu.Lock()
	server := s.server
	s.server = nil
	s.listener = nil
	s.url = ""
	s.mu.Unlock()
	if server == nil {
		return nil
	}
	stopCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return server.Shutdown(stopCtx)
}

func (s *Server) Status() bookmark.ServerStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.statusLocked()
}

func (s *Server) statusLocked() bookmark.ServerStatus {
	if s.server == nil {
		return bookmark.ServerStatus{}
	}
	addr := ""
	if s.listener != nil {
		addr = s.listener.Addr().String()
	}
	return bookmark.ServerStatus{Running: true, URL: s.url, Addr: addr}
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	items, err := s.store.AllBookmarks(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html, err := exporter.RenderCatalogHTML(exporter.BuildCatalog("Cola Bookmarks", items), "classic")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(html))
}

func (s *Server) handleBookmarks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	req := bookmark.SearchRequest{
		Query:  r.URL.Query().Get("q"),
		Folder: r.URL.Query().Get("folder"),
		Tags:   splitQueryList(r.URL.Query().Get("tags")),
		Sort:   r.URL.Query().Get("sort"),
	}
	if limit := r.URL.Query().Get("limit"); limit != "" {
		req.Limit, _ = strconv.Atoi(limit)
	}
	if offset := r.URL.Query().Get("offset"); offset != "" {
		req.Offset, _ = strconv.Atoi(offset)
	}
	result, err := s.store.ListBookmarks(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, result)
}

func (s *Server) handleMeta(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	folders, err := s.store.ListFolders(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tags, err := s.store.ListTags(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{"folders": folders, "tags": tags})
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(true)
	_ = encoder.Encode(value)
}

func splitQueryList(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'self' 'unsafe-inline'; img-src 'self' data: https:; connect-src 'self'")
		next.ServeHTTP(w, r)
	})
}
