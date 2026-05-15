// thumbnail 相关逻辑集中处理两路缓存：自动获取缩略图与用户自定义缩略图。
package storage

import (
	"bytes"
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"cola/internal/bookmark"
)

const (
	maxThumbnailBytes = 8 << 20
	thumbnailMaxWidth = 640
)

var (
	autoDownloadSlots   = make(chan struct{}, 2)
	screenshotSlots     = make(chan struct{}, 1)
	thumbnailHTTPClient = &http.Client{Timeout: 14 * time.Second}
)

type thumbnailColumn struct {
	name       string
	definition string
}

type thumbnailAsset struct {
	source      string
	filePath    string
	thumbPath   string
	originalURL string
	mime        string
	width       int
	height      int
	size        int64
}

var thumbnailColumns = []thumbnailColumn{
	{name: "use_auto_thumbnail", definition: "INTEGER NOT NULL DEFAULT 1"},
	{name: "auto_source", definition: "TEXT NOT NULL DEFAULT ''"},
	{name: "auto_file_path", definition: "TEXT NOT NULL DEFAULT ''"},
	{name: "auto_thumb_path", definition: "TEXT NOT NULL DEFAULT ''"},
	{name: "auto_original_url", definition: "TEXT NOT NULL DEFAULT ''"},
	{name: "auto_mime", definition: "TEXT NOT NULL DEFAULT ''"},
	{name: "auto_width", definition: "INTEGER NOT NULL DEFAULT 0"},
	{name: "auto_height", definition: "INTEGER NOT NULL DEFAULT 0"},
	{name: "auto_size", definition: "INTEGER NOT NULL DEFAULT 0"},
	{name: "auto_status", definition: "TEXT NOT NULL DEFAULT ''"},
	{name: "auto_error", definition: "TEXT NOT NULL DEFAULT ''"},
	{name: "auto_fetched_at", definition: "TEXT"},
	{name: "custom_source", definition: "TEXT NOT NULL DEFAULT ''"},
	{name: "custom_file_path", definition: "TEXT NOT NULL DEFAULT ''"},
	{name: "custom_thumb_path", definition: "TEXT NOT NULL DEFAULT ''"},
	{name: "custom_original_url", definition: "TEXT NOT NULL DEFAULT ''"},
	{name: "custom_mime", definition: "TEXT NOT NULL DEFAULT ''"},
	{name: "custom_width", definition: "INTEGER NOT NULL DEFAULT 0"},
	{name: "custom_height", definition: "INTEGER NOT NULL DEFAULT 0"},
	{name: "custom_size", definition: "INTEGER NOT NULL DEFAULT 0"},
	{name: "custom_created_at", definition: "TEXT"},
}

func (s *Store) migrateLegacyPreviewRows(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `UPDATE bookmark_previews SET
		use_auto_thumbnail = 1,
		auto_source = CASE WHEN COALESCE(auto_source, '') = '' THEN source ELSE auto_source END,
		auto_file_path = CASE WHEN COALESCE(auto_file_path, '') = '' THEN file_path ELSE auto_file_path END,
		auto_thumb_path = CASE WHEN COALESCE(auto_thumb_path, '') = '' THEN thumb_path ELSE auto_thumb_path END,
		auto_original_url = CASE WHEN COALESCE(auto_original_url, '') = '' THEN original_url ELSE auto_original_url END,
		auto_mime = CASE WHEN COALESCE(auto_mime, '') = '' THEN mime ELSE auto_mime END,
		auto_width = CASE WHEN COALESCE(auto_width, 0) = 0 THEN width ELSE auto_width END,
		auto_height = CASE WHEN COALESCE(auto_height, 0) = 0 THEN height ELSE auto_height END,
		auto_size = CASE WHEN COALESCE(auto_size, 0) = 0 THEN size ELSE auto_size END,
		auto_status = CASE WHEN COALESCE(auto_status, '') = '' THEN 'ready' ELSE auto_status END,
		auto_fetched_at = CASE WHEN COALESCE(auto_fetched_at, '') = '' THEN created_at ELSE auto_fetched_at END
		WHERE COALESCE(file_path, '') != '' AND COALESCE(auto_file_path, '') = ''`)
	return err
}

func (s *Store) SaveCustomThumbnail(ctx context.Context, bookmarkID string, input bookmark.ThumbnailUploadInput) (bookmark.Thumbnail, error) {
	if _, err := s.GetBookmark(ctx, bookmarkID); err != nil {
		return bookmark.Thumbnail{}, err
	}
	data, mimeType, err := decodeThumbnailUpload(input)
	if err != nil {
		return bookmark.Thumbnail{}, err
	}
	asset, err := s.cacheThumbnailBytes(data, input.FileName, mimeType, "upload", "")
	if err != nil {
		return bookmark.Thumbnail{}, err
	}
	if err := s.ensureThumbnailRow(ctx, bookmarkID); err != nil {
		return bookmark.Thumbnail{}, err
	}
	now := time.Now().UTC()
	if _, err := s.db.ExecContext(ctx, `UPDATE bookmark_previews SET
		use_auto_thumbnail = 0,
		custom_source = ?, custom_file_path = ?, custom_thumb_path = ?, custom_original_url = ?,
		custom_mime = ?, custom_width = ?, custom_height = ?, custom_size = ?, custom_created_at = ?
		WHERE bookmark_id = ?`,
		asset.source, asset.filePath, asset.thumbPath, asset.originalURL, asset.mime, asset.width, asset.height, asset.size, formatTime(now), bookmarkID); err != nil {
		return bookmark.Thumbnail{}, err
	}
	return s.thumbnailByBookmarkID(ctx, bookmarkID)
}

func (s *Store) SaveCustomThumbnailURL(ctx context.Context, bookmarkID string, input bookmark.ThumbnailURLInput) (bookmark.Thumbnail, error) {
	if _, err := s.GetBookmark(ctx, bookmarkID); err != nil {
		return bookmark.Thumbnail{}, err
	}
	asset, err := s.downloadThumbnailAsset(ctx, strings.TrimSpace(input.URL), "remote")
	if err != nil {
		return bookmark.Thumbnail{}, err
	}
	if err := s.ensureThumbnailRow(ctx, bookmarkID); err != nil {
		return bookmark.Thumbnail{}, err
	}
	now := time.Now().UTC()
	if _, err := s.db.ExecContext(ctx, `UPDATE bookmark_previews SET
		use_auto_thumbnail = 0,
		custom_source = ?, custom_file_path = ?, custom_thumb_path = ?, custom_original_url = ?,
		custom_mime = ?, custom_width = ?, custom_height = ?, custom_size = ?, custom_created_at = ?
		WHERE bookmark_id = ?`,
		asset.source, asset.filePath, asset.thumbPath, asset.originalURL, asset.mime, asset.width, asset.height, asset.size, formatTime(now), bookmarkID); err != nil {
		return bookmark.Thumbnail{}, err
	}
	return s.thumbnailByBookmarkID(ctx, bookmarkID)
}

func (s *Store) SetThumbnailAutoMode(ctx context.Context, bookmarkID string, input bookmark.ThumbnailModeInput) (bookmark.Thumbnail, error) {
	if _, err := s.GetBookmark(ctx, bookmarkID); err != nil {
		return bookmark.Thumbnail{}, err
	}
	if err := s.ensureThumbnailRow(ctx, bookmarkID); err != nil {
		return bookmark.Thumbnail{}, err
	}
	useAuto := 0
	if input.UseAuto {
		useAuto = 1
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE bookmark_previews SET use_auto_thumbnail = ? WHERE bookmark_id = ?`, useAuto, bookmarkID); err != nil {
		return bookmark.Thumbnail{}, err
	}
	return s.thumbnailByBookmarkID(ctx, bookmarkID)
}

func (s *Store) ClearCustomThumbnail(ctx context.Context, bookmarkID string) (bookmark.Thumbnail, error) {
	if _, err := s.GetBookmark(ctx, bookmarkID); err != nil {
		return bookmark.Thumbnail{}, err
	}
	if err := s.ensureThumbnailRow(ctx, bookmarkID); err != nil {
		return bookmark.Thumbnail{}, err
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE bookmark_previews SET
		use_auto_thumbnail = 1,
		custom_source = '', custom_file_path = '', custom_thumb_path = '', custom_original_url = '',
		custom_mime = '', custom_width = 0, custom_height = 0, custom_size = 0, custom_created_at = NULL
		WHERE bookmark_id = ?`, bookmarkID); err != nil {
		return bookmark.Thumbnail{}, err
	}
	return s.thumbnailByBookmarkID(ctx, bookmarkID)
}

func (s *Store) RefreshAutoThumbnail(ctx context.Context, bookmarkID string) (bookmark.Thumbnail, error) {
	item, err := s.GetBookmark(ctx, bookmarkID)
	if err != nil {
		return bookmark.Thumbnail{}, err
	}
	if err := s.ensureThumbnailRow(ctx, bookmarkID); err != nil {
		return bookmark.Thumbnail{}, err
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE bookmark_previews SET auto_status = 'loading', auto_error = '' WHERE bookmark_id = ?`, bookmarkID); err != nil {
		return bookmark.Thumbnail{}, err
	}
	asset, err := s.fetchAutoThumbnailAsset(ctx, item)
	if err != nil {
		_, _ = s.db.ExecContext(ctx, `UPDATE bookmark_previews SET auto_status = 'error', auto_error = ? WHERE bookmark_id = ?`, err.Error(), bookmarkID)
		return s.thumbnailByBookmarkID(ctx, bookmarkID)
	}
	now := time.Now().UTC()
	if _, err := s.db.ExecContext(ctx, `UPDATE bookmark_previews SET
		auto_source = ?, auto_file_path = ?, auto_thumb_path = ?, auto_original_url = ?,
		auto_mime = ?, auto_width = ?, auto_height = ?, auto_size = ?, auto_status = 'ready', auto_error = '', auto_fetched_at = ?
		WHERE bookmark_id = ?`,
		asset.source, asset.filePath, asset.thumbPath, asset.originalURL, asset.mime, asset.width, asset.height, asset.size, formatTime(now), bookmarkID); err != nil {
		return bookmark.Thumbnail{}, err
	}
	return s.thumbnailByBookmarkID(ctx, bookmarkID)
}

func (s *Store) EnsureAutoThumbnail(ctx context.Context, bookmarkID string) (bookmark.Thumbnail, error) {
	if _, err := s.GetBookmark(ctx, bookmarkID); err != nil {
		return bookmark.Thumbnail{}, err
	}
	thumbnail, err := s.loadThumbnail(ctx, bookmarkID)
	if err != nil {
		return bookmark.Thumbnail{}, err
	}
	if thumbnail != nil {
		if !thumbnail.UseAuto || thumbnailHasAutoAsset(thumbnail) || thumbnail.AutoStatus == "error" {
			return *thumbnail, nil
		}
	}
	return s.RefreshAutoThumbnail(ctx, bookmarkID)
}

func (s *Store) ensureThumbnailRow(ctx context.Context, bookmarkID string) error {
	now := formatTime(time.Now().UTC())
	_, err := s.db.ExecContext(ctx, `INSERT INTO bookmark_previews
		(id, bookmark_id, source, file_path, thumb_path, original_url, mime, width, height, size, created_at, use_auto_thumbnail)
		VALUES(?, ?, '', '', '', '', '', 0, 0, 0, ?, 1)
		ON CONFLICT(bookmark_id) DO NOTHING`, newID("thumb"), bookmarkID, now)
	return err
}

func (s *Store) thumbnailByBookmarkID(ctx context.Context, bookmarkID string) (bookmark.Thumbnail, error) {
	thumbnail, err := s.loadThumbnail(ctx, bookmarkID)
	if err != nil {
		return bookmark.Thumbnail{}, err
	}
	if thumbnail == nil {
		return bookmark.Thumbnail{}, sql.ErrNoRows
	}
	return *thumbnail, nil
}

func (s *Store) loadThumbnail(ctx context.Context, bookmarkID string) (*bookmark.Thumbnail, error) {
	var thumbnail bookmark.Thumbnail
	var useAuto int
	var autoFetchedAt, customCreatedAt sql.NullString
	err := s.db.QueryRowContext(ctx, `SELECT id, bookmark_id, use_auto_thumbnail,
		auto_source, auto_file_path, auto_thumb_path, auto_original_url, auto_mime, auto_width, auto_height, auto_size, auto_status, auto_error, auto_fetched_at,
		custom_source, custom_file_path, custom_thumb_path, custom_original_url, custom_mime, custom_width, custom_height, custom_size, custom_created_at
		FROM bookmark_previews WHERE bookmark_id = ?`, bookmarkID).Scan(
		&thumbnail.ID, &thumbnail.BookmarkID, &useAuto,
		&thumbnail.AutoSource, &thumbnail.AutoFilePath, &thumbnail.AutoThumbPath, &thumbnail.AutoOriginalURL, &thumbnail.AutoMime,
		&thumbnail.AutoWidth, &thumbnail.AutoHeight, &thumbnail.AutoSize, &thumbnail.AutoStatus, &thumbnail.AutoError, &autoFetchedAt,
		&thumbnail.CustomSource, &thumbnail.CustomFilePath, &thumbnail.CustomThumbPath, &thumbnail.CustomOriginalURL, &thumbnail.CustomMime,
		&thumbnail.CustomWidth, &thumbnail.CustomHeight, &thumbnail.CustomSize, &customCreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	thumbnail.UseAuto = useAuto == 1
	if autoFetchedAt.Valid && autoFetchedAt.String != "" {
		parsed := parseTime(autoFetchedAt.String)
		thumbnail.AutoFetchedAt = &parsed
	}
	if customCreatedAt.Valid && customCreatedAt.String != "" {
		parsed := parseTime(customCreatedAt.String)
		thumbnail.CustomCreatedAt = &parsed
	}
	thumbnail.AutoFilePath = previewAssetURL(thumbnail.AutoFilePath)
	thumbnail.AutoThumbPath = previewAssetURL(thumbnail.AutoThumbPath)
	thumbnail.CustomFilePath = previewAssetURL(thumbnail.CustomFilePath)
	thumbnail.CustomThumbPath = previewAssetURL(thumbnail.CustomThumbPath)
	if thumbnail.UseAuto {
		thumbnail.DisplayPath = firstNonEmpty(thumbnail.AutoThumbPath, thumbnail.AutoFilePath)
		thumbnail.DisplaySource = thumbnail.AutoSource
	} else {
		thumbnail.DisplayPath = firstNonEmpty(thumbnail.CustomThumbPath, thumbnail.CustomFilePath)
		thumbnail.DisplaySource = thumbnail.CustomSource
	}
	return &thumbnail, nil
}

func (s *Store) fetchAutoThumbnailAsset(ctx context.Context, item bookmark.Bookmark) (thumbnailAsset, error) {
	html, htmlErr := s.fetchBookmarkHTML(ctx, item.URL)
	if htmlErr == nil {
		if imageURL := extractOpenGraphImage(html); imageURL != "" {
			if resolved, err := resolveURL(item.URL, imageURL); err == nil {
				if asset, err := s.downloadThumbnailAsset(ctx, resolved, "og"); err == nil {
					return asset, nil
				}
			}
		}
		if faviconURL := extractFaviconURL(html); faviconURL != "" {
			if resolved, err := resolveURL(item.URL, faviconURL); err == nil {
				if asset, err := s.downloadThumbnailAsset(ctx, resolved, "favicon"); err == nil {
					return asset, nil
				}
			}
		}
	}
	if fallback, err := resolveURL(item.URL, "/favicon.ico"); err == nil {
		if asset, err := s.downloadThumbnailAsset(ctx, fallback, "favicon"); err == nil {
			return asset, nil
		}
	}
	if asset, err := s.captureScreenshotAsset(ctx, item.URL); err == nil {
		return asset, nil
	}
	if htmlErr != nil {
		return thumbnailAsset{}, htmlErr
	}
	return thumbnailAsset{}, errors.New("thumbnail image not found")
}

func (s *Store) fetchBookmarkHTML(ctx context.Context, rawURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := thumbnailHTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("thumbnail page status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (s *Store) downloadThumbnailAsset(ctx context.Context, rawURL, source string) (thumbnailAsset, error) {
	normalizedURL, err := normalizeThumbnailRemoteURL(rawURL)
	if err != nil {
		return thumbnailAsset{}, err
	}
	release, err := acquireThumbnailSlot(ctx, autoDownloadSlots)
	if err != nil {
		return thumbnailAsset{}, err
	}
	defer release()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, normalizedURL, nil)
	if err != nil {
		return thumbnailAsset{}, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; ColaBookmarks/1.0)")
	resp, err := thumbnailHTTPClient.Do(req)
	if err != nil {
		return thumbnailAsset{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return thumbnailAsset{}, fmt.Errorf("thumbnail image status %d", resp.StatusCode)
	}
	data, err := readLimitedBytes(resp.Body, maxThumbnailBytes)
	if err != nil {
		return thumbnailAsset{}, err
	}
	contentType := strings.Split(resp.Header.Get("Content-Type"), ";")[0]
	return s.cacheThumbnailBytes(data, filepath.Base(resp.Request.URL.Path), contentType, source, normalizedURL)
}

func (s *Store) captureScreenshotAsset(ctx context.Context, rawURL string) (thumbnailAsset, error) {
	workerPath, err := thumbnailWorkerPath()
	if err != nil {
		return thumbnailAsset{}, err
	}
	release, err := acquireThumbnailSlot(ctx, screenshotSlots)
	if err != nil {
		return thumbnailAsset{}, err
	}
	defer release()
	tempFile, err := os.CreateTemp(s.dataDir, "thumbnail-shot-*.png")
	if err != nil {
		return thumbnailAsset{}, err
	}
	tempPath := tempFile.Name()
	_ = tempFile.Close()
	defer os.Remove(tempPath)
	nodePath := strings.TrimSpace(os.Getenv("COLA_NODE_PATH"))
	if nodePath == "" {
		nodePath = "node"
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, nodePath, workerPath, rawURL, tempPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return thumbnailAsset{}, fmt.Errorf("thumbnail screenshot failed: %s", strings.TrimSpace(string(output)))
	}
	data, err := os.ReadFile(tempPath)
	if err != nil {
		return thumbnailAsset{}, err
	}
	return s.cacheThumbnailBytes(data, "screenshot.png", "image/png", "screenshot", rawURL)
}

func (s *Store) cacheThumbnailBytes(data []byte, fileName, mimeType, source, originalURL string) (thumbnailAsset, error) {
	if len(data) == 0 {
		return thumbnailAsset{}, errors.New("thumbnail image is empty")
	}
	if len(data) > maxThumbnailBytes {
		return thumbnailAsset{}, errors.New("thumbnail image is too large")
	}
	mimeType = normalizeImageMIME(mimeType, data)
	if !isSupportedImageMIME(mimeType) {
		return thumbnailAsset{}, errors.New("thumbnail image type is not supported")
	}
	ext := imageExtension(fileName, mimeType)
	hash := sha1.Sum(data)
	hashName := fmt.Sprintf("%x", hash[:])
	originalRel := filepath.ToSlash(filepath.Join("previews", "original", hashName+ext))
	originalPath := filepath.Join(s.dataDir, filepath.FromSlash(originalRel))
	if err := os.MkdirAll(filepath.Dir(originalPath), 0o755); err != nil {
		return thumbnailAsset{}, err
	}
	if err := os.WriteFile(originalPath, data, 0o644); err != nil {
		return thumbnailAsset{}, err
	}
	width, height := readImageSize(data)
	thumbRel := originalRel
	if thumbData, ok := thumbnailJPEG(data); ok {
		thumbRel = filepath.ToSlash(filepath.Join("previews", "thumb", hashName+".jpg"))
		thumbPath := filepath.Join(s.dataDir, filepath.FromSlash(thumbRel))
		if err := os.MkdirAll(filepath.Dir(thumbPath), 0o755); err != nil {
			return thumbnailAsset{}, err
		}
		if err := os.WriteFile(thumbPath, thumbData, 0o644); err != nil {
			return thumbnailAsset{}, err
		}
	}
	return thumbnailAsset{
		source:      source,
		filePath:    originalRel,
		thumbPath:   thumbRel,
		originalURL: strings.TrimSpace(originalURL),
		mime:        mimeType,
		width:       width,
		height:      height,
		size:        int64(len(data)),
	}, nil
}

func decodeThumbnailUpload(input bookmark.ThumbnailUploadInput) ([]byte, string, error) {
	raw := strings.TrimSpace(input.Data)
	mimeType := strings.TrimSpace(input.Mime)
	if comma := strings.Index(raw, ","); strings.HasPrefix(raw, "data:") && comma >= 0 {
		header := raw[:comma]
		raw = raw[comma+1:]
		if semi := strings.Index(header, ";"); semi > len("data:") {
			mimeType = header[len("data:"):semi]
		}
	}
	data, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, "", err
	}
	return data, mimeType, nil
}

func dataURLFromBytes(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func readLimitedBytes(reader io.Reader, limit int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, errors.New("thumbnail image is too large")
	}
	return data, nil
}

func normalizeThumbnailRemoteURL(rawURL string) (string, error) {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(rawURL))
	if err != nil {
		return "", err
	}
	switch strings.ToLower(parsed.Scheme) {
	case "http", "https":
		return parsed.String(), nil
	default:
		return "", errors.New("thumbnail url must use http or https")
	}
}

func acquireThumbnailSlot(ctx context.Context, slots chan struct{}) (func(), error) {
	// 下载与截图共享轻量队列，避免列表懒加载时同时占满网络和浏览器资源。
	select {
	case slots <- struct{}{}:
		return func() {
			<-slots
		}, nil
	case <-ctx.Done():
		return func() {}, ctx.Err()
	}
}

func normalizeImageMIME(mimeType string, data []byte) string {
	mimeType = strings.ToLower(strings.TrimSpace(strings.Split(mimeType, ";")[0]))
	if mimeType == "image/jpg" {
		mimeType = "image/jpeg"
	}
	if detected := detectImageMIME(data); detected != "" {
		return detected
	}
	return http.DetectContentType(data)
}

func detectImageMIME(data []byte) string {
	contentType := http.DetectContentType(data)
	switch contentType {
	case "image/jpeg", "image/png", "image/gif", "image/webp":
		return contentType
	}
	if looksLikeWEBP(data) {
		return "image/webp"
	}
	if looksLikeICO(data) {
		return "image/x-icon"
	}
	if looksLikeSVG(data) {
		return "image/svg+xml"
	}
	return ""
}

func looksLikeWEBP(data []byte) bool {
	return len(data) >= 12 && string(data[:4]) == "RIFF" && string(data[8:12]) == "WEBP"
}

func looksLikeICO(data []byte) bool {
	return len(data) >= 4 && data[0] == 0 && data[1] == 0 && (data[2] == 1 || data[2] == 2) && data[3] == 0
}

func looksLikeSVG(data []byte) bool {
	sampleLength := len(data)
	if sampleLength > 512 {
		sampleLength = 512
	}
	sample := strings.ToLower(strings.TrimSpace(string(data[:sampleLength])))
	return strings.Contains(sample, "<svg")
}

func isSupportedImageMIME(mimeType string) bool {
	switch mimeType {
	case "image/jpeg", "image/png", "image/gif", "image/webp", "image/x-icon", "image/vnd.microsoft.icon", "image/svg+xml":
		return true
	default:
		return false
	}
}

func imageExtension(fileName, mimeType string) string {
	switch mimeType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "image/svg+xml":
		return ".svg"
	}
	if exts, _ := mime.ExtensionsByType(mimeType); len(exts) > 0 {
		if exts[0] == ".jpe" {
			return ".jpg"
		}
		return exts[0]
	}
	if ext := strings.ToLower(filepath.Ext(fileName)); ext != "" {
		return ext
	}
	return ".img"
}

func readImageSize(data []byte) (int, int) {
	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return 0, 0
	}
	return config.Width, config.Height
}

func thumbnailJPEG(data []byte) ([]byte, bool) {
	source, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, false
	}
	bounds := source.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width <= 0 || height <= 0 {
		return nil, false
	}
	nextWidth := thumbnailMaxWidth
	if width < thumbnailMaxWidth {
		nextWidth = width
	}
	nextHeight := int(float64(height) * (float64(nextWidth) / float64(width)))
	if nextHeight <= 0 {
		nextHeight = 1
	}
	target := image.NewRGBA(image.Rect(0, 0, nextWidth, nextHeight))
	draw.Draw(target, target.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
	for y := 0; y < nextHeight; y++ {
		for x := 0; x < nextWidth; x++ {
			srcX := bounds.Min.X + x*width/nextWidth
			srcY := bounds.Min.Y + y*height/nextHeight
			target.Set(x, y, source.At(srcX, srcY))
		}
	}
	var out bytes.Buffer
	if err := jpeg.Encode(&out, target, &jpeg.Options{Quality: 78}); err != nil {
		return nil, false
	}
	return out.Bytes(), true
}

func extractFaviconURL(html string) string {
	linkPattern := regexp.MustCompile(`(?is)<link[^>]+>`)
	for _, tag := range linkPattern.FindAllString(html, -1) {
		rel := strings.ToLower(extractHTMLAttr(tag, "rel"))
		if !strings.Contains(rel, "icon") {
			continue
		}
		if href := strings.TrimSpace(extractHTMLAttr(tag, "href")); href != "" {
			return href
		}
	}
	return ""
}

func extractHTMLAttr(tag, attr string) string {
	pattern := regexp.MustCompile(`(?is)\s` + regexp.QuoteMeta(attr) + `\s*=\s*["']([^"']+)["']`)
	if match := pattern.FindStringSubmatch(tag); len(match) == 2 {
		return strings.TrimSpace(match[1])
	}
	return ""
}

func thumbnailWorkerPath() (string, error) {
	candidates := []string{}
	if configured := strings.TrimSpace(os.Getenv("COLA_THUMBNAIL_WORKER")); configured != "" {
		candidates = append(candidates, configured)
	}
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates,
			filepath.Join(cwd, "frontend", "scripts", "thumbnail-worker.mjs"),
			filepath.Join(cwd, "scripts", "thumbnail-worker.mjs"),
		)
	}
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(dir, "frontend", "scripts", "thumbnail-worker.mjs"),
			filepath.Join(dir, "scripts", "thumbnail-worker.mjs"),
		)
	}
	for _, candidate := range candidates {
		if stat, err := os.Stat(candidate); err == nil && !stat.IsDir() {
			return candidate, nil
		}
	}
	return "", errors.New("thumbnail screenshot worker is not available")
}

func previewFromThumbnail(thumbnail bookmark.Thumbnail) bookmark.Preview {
	filePath := thumbnail.AutoFilePath
	thumbPath := thumbnail.AutoThumbPath
	source := thumbnail.AutoSource
	originalURL := thumbnail.AutoOriginalURL
	mimeType := thumbnail.AutoMime
	width := thumbnail.AutoWidth
	height := thumbnail.AutoHeight
	size := thumbnail.AutoSize
	if !thumbnail.UseAuto {
		filePath = thumbnail.CustomFilePath
		thumbPath = thumbnail.CustomThumbPath
		source = thumbnail.CustomSource
		originalURL = thumbnail.CustomOriginalURL
		mimeType = thumbnail.CustomMime
		width = thumbnail.CustomWidth
		height = thumbnail.CustomHeight
		size = thumbnail.CustomSize
	} else if source != "" {
		source = "auto"
	}
	return bookmark.Preview{
		ID:          thumbnail.ID,
		BookmarkID:  thumbnail.BookmarkID,
		Source:      source,
		FilePath:    filePath,
		ThumbPath:   thumbPath,
		OriginalURL: originalURL,
		Mime:        mimeType,
		Width:       width,
		Height:      height,
		Size:        size,
	}
}

func thumbnailHasAutoAsset(thumbnail *bookmark.Thumbnail) bool {
	if thumbnail == nil {
		return false
	}
	return firstNonEmpty(thumbnail.AutoThumbPath, thumbnail.AutoFilePath) != ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
