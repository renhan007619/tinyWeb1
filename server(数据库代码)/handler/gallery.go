// Package handler 图库功能 - Day 3 完整实现
// 核心功能：上传、查询、删除（软删除）、专辑管理、批量操作、回收站
package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"tinyweb1/db"
	"tinyweb1/middleware"
	"tinyweb1/model"
)

const GalleryUploadDir = "uploads/gallery"

var AllowedImageTypes = map[string]bool{
	"image/jpeg": true, "image/jpg": true, "image/png": true,
	"image/gif": true, "image/webp": true,
}

const MaxImageSize = 50 * 1024 * 1024 // 50MB

// getGalleryUserID 从JWT context获取用户ID
func getGalleryUserID(r *http.Request) uint {
	id, _ := middleware.GetUserID(r.Context())
	return id
}

// ============================================
// POST /api/gallery/upload - 上传图片
// ============================================
func UploadImage(w http.ResponseWriter, r *http.Request) {
	userID := getGalleryUserID(r)
	if userID == 0 {
		sendJSON(w, http.StatusUnauthorized, model.ErrorResponse(401, "未登录"))
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "解析表单失败"))
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "请选择文件"))
		return
	}

	description := r.FormValue("description")
	tags := r.FormValue("tags")

	today := time.Now().Format("2006-01-02")
	userDir := filepath.Join(GalleryUploadDir, fmt.Sprintf("user_%d", userID), today)
	os.MkdirAll(userDir, 0755)

	var uploaded []model.UploadImageResponse
	var errors []string

	for _, file := range files {
		resp, err := processUpload(file, userID, userDir, today, description, tags)
		if err != nil {
			errors = append(errors, file.Filename+": "+err.Error())
			continue
		}
		uploaded = append(uploaded, *resp)
	}

	result := map[string]interface{}{
		"success": uploaded, "failed": len(errors), "errors": errors,
	}
	if len(uploaded) == 0 {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "全部失败"))
		return
	}
	sendJSON(w, http.StatusOK, model.SuccessResponse(result))
}

func processUpload(file *multipart.FileHeader, userID uint, userDir, today, desc, tags string) (*model.UploadImageResponse, error) {
	if file.Size > MaxImageSize {
		return nil, fmt.Errorf("文件太大")
	}

	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	// 检测文件类型
	buf := make([]byte, 512)
	n, _ := src.Read(buf)
	src.Seek(0, io.SeekStart)
	mimeType := http.DetectContentType(buf[:n])
	if !AllowedImageTypes[mimeType] {
		return nil, fmt.Errorf("不支持的类型")
	}

	// 生成文件名
	ext := filepath.Ext(file.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	newName := fmt.Sprintf("%s_%s%s", generateID(), sanitizeName(file.Filename), ext)
	filePath := filepath.Join(userDir, newName)

	// 保存文件
	dst, _ := os.Create(filePath)
	defer dst.Close()
	io.Copy(dst, src)

	// 保存到数据库
	img := model.GalleryImage{
		UserID: userID, FilePath: filePath, FileName: file.Filename,
		FileSize: file.Size, MimeType: mimeType, UploadDate: today,
		Description: desc, Tags: tags,
	}
	if err := db.GetDB().Create(&img).Error; err != nil {
		os.Remove(filePath)
		return nil, err
	}

	return &model.UploadImageResponse{ID: img.ID, FilePath: filePath, UploadDate: today}, nil
}

func sanitizeName(name string) string {
	name = filepath.Base(name)
	name = strings.TrimSuffix(name, filepath.Ext(name))
	r := strings.NewReplacer(" ", "_", "..", "", "/", "", "\\", "", ":", "", "*", "", "?", "", "\"", "", "<", "", ">", "", "|", "")
	return r.Replace(name)
}

// ============================================
// GET /api/gallery/images - 图片列表
// ============================================
func GetImageList(w http.ResponseWriter, r *http.Request) {
	userID := getGalleryUserID(r)
	if userID == 0 {
		sendJSON(w, http.StatusUnauthorized, model.ErrorResponse(401, "未登录"))
		return
	}

	page, pageSize := 1, 50
	date := r.URL.Query().Get("date")
	fmt.Sscanf(r.URL.Query().Get("page"), "%d", &page)
	fmt.Sscanf(r.URL.Query().Get("page_size"), "%d", &pageSize)
	if pageSize > 100 {
		pageSize = 100
	}

	query := db.GetDB().Model(&model.GalleryImage{}).
		Where("user_id = ? AND is_deleted = ?", userID, false)
	if date != "" {
		query = query.Where("upload_date = ?", date)
	}

	var total int64
	query.Count(&total)

	var images []model.GalleryImage
	query.Order("created_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&images)

	items := make([]model.GalleryImageItem, len(images))
	for i, img := range images {
		items[i] = convertItem(img)
	}

	sendJSON(w, http.StatusOK, model.SuccessResponse(model.ImageListResponse{
		List: items, Total: total, Page: page, PageSize: pageSize,
		TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
	}))
}

// ============================================
// GET /api/gallery/images/by-date - 按日期分组
// ============================================
func GetImagesByDate(w http.ResponseWriter, r *http.Request) {
	userID := getGalleryUserID(r)
	if userID == 0 {
		sendJSON(w, http.StatusUnauthorized, model.ErrorResponse(401, "未登录"))
		return
	}

	var dates []string
	db.GetDB().Model(&model.GalleryImage{}).
		Select("DISTINCT upload_date").
		Where("user_id = ? AND is_deleted = ?", userID, false).
		Order("upload_date DESC").Pluck("upload_date", &dates)

	var groups []model.DateGroupItem
	for _, d := range dates {
		var imgs []model.GalleryImage
		db.GetDB().Where("user_id = ? AND upload_date = ? AND is_deleted = ?", userID, d, false).
			Order("created_at DESC").Find(&imgs)

		items := make([]model.GalleryImageItem, len(imgs))
		for i, img := range imgs {
			items[i] = convertItem(img)
		}
		groups = append(groups, model.DateGroupItem{Date: d, Count: int64(len(imgs)), Images: items})
	}

	sendJSON(w, http.StatusOK, model.SuccessResponse(model.DateGroupResponse{
		Groups: groups, Total: int64(len(dates)), TotalDates: len(groups),
	}))
}

// ============================================
// DELETE /api/gallery/images/:id - 删除（软删除）
// ============================================
func DeleteImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		sendJSON(w, 405, model.ErrorResponse(405, "方法不允许"))
		return
	}
	userID := getGalleryUserID(r)
	if userID == 0 {
		sendJSON(w, 401, model.ErrorResponse(401, "未登录"))
		return
	}

	id := extractID(r.URL.Path, "/api/gallery/images/")
	now := time.Now()

	result := db.GetDB().Model(&model.GalleryImage{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(map[string]interface{}{"is_deleted": true, "deleted_time": now})

	if result.RowsAffected == 0 {
		sendJSON(w, 404, model.ErrorResponse(404, "图片不存在"))
		return
	}
	sendJSON(w, 200, model.SuccessResponse(map[string]string{"message": "已移入回收站"}))
}

// ============================================
// PUT /api/gallery/images/:id - 更新信息
// ============================================
func UpdateImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		sendJSON(w, 405, model.ErrorResponse(405, "方法不允许"))
		return
	}
	userID := getGalleryUserID(r)
	id := extractID(r.URL.Path, "/api/gallery/images/")

	var req model.UpdateImageRequest
	json.NewDecoder(r.Body).Decode(&req)

	updates := map[string]interface{}{}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Tags != nil {
		updates["tags"] = *req.Tags
	}
	if req.IsFavorite != nil {
		updates["is_favorite"] = *req.IsFavorite
	}

	if len(updates) == 0 {
		sendJSON(w, 400, model.ErrorResponse(400, "无更新内容"))
		return
	}

	result := db.GetDB().Model(&model.GalleryImage{}).
		Where("id = ? AND user_id = ?", id, userID).Updates(updates)

	if result.RowsAffected == 0 {
		sendJSON(w, 404, model.ErrorResponse(404, "图片不存在"))
		return
	}
	sendJSON(w, 200, model.SuccessResponse(map[string]string{"message": "更新成功"}))
}

// ============================================
// 专辑相关 API（简化版）
// ============================================

// GetAlbumList GET /api/gallery/albums
func GetAlbumList(w http.ResponseWriter, r *http.Request) {
	userID := getGalleryUserID(r)
	var albums []model.GalleryAlbum
	db.GetDB().Where("user_id = ?", userID).Order("created_at DESC").Find(&albums)

	resp := make([]model.AlbumResponse, len(albums))
	for i, a := range albums {
		var count int64
		db.GetDB().Model(&model.GalleryImage{}).Where("album_id = ?", a.ID).Count(&count)
		resp[i] = model.AlbumResponse{
			ID: a.ID, Name: a.Name, Description: a.Description,
			ImageCount: count, CreatedAt: a.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	sendJSON(w, 200, model.SuccessResponse(resp))
}

// CreateAlbum POST /api/gallery/albums
func CreateAlbum(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, 405, model.ErrorResponse(405, "方法不允许"))
		return
	}
	userID := getGalleryUserID(r)
	var req model.CreateAlbumRequest
	json.NewDecoder(r.Body).Decode(&req)

	if strings.TrimSpace(req.Name) == "" {
		sendJSON(w, 400, model.ErrorResponse(400, "名称不能为空"))
		return
	}

	album := model.GalleryAlbum{UserID: userID, Name: req.Name, Description: req.Description}
	db.GetDB().Create(&album)

	sendJSON(w, 201, model.SuccessResponse(model.AlbumResponse{
		ID: album.ID, Name: album.Name, CreatedAt: album.CreatedAt.Format("2006-01-02 15:04:05"),
	}))
}

// UpdateAlbum PUT /api/gallery/albums/:id
func UpdateAlbum(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		sendJSON(w, 405, model.ErrorResponse(405, "方法不允许"))
		return
	}
	userID := getGalleryUserID(r)
	id := extractID(r.URL.Path, "/api/gallery/albums/")

	var req model.UpdateAlbumRequest
	json.NewDecoder(r.Body).Decode(&req)

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = strings.TrimSpace(req.Name)
	}
	updates["description"] = req.Description

	result := db.GetDB().Model(&model.GalleryAlbum{}).
		Where("id = ? AND user_id = ?", id, userID).Updates(updates)

	if result.RowsAffected == 0 {
		sendJSON(w, 404, model.ErrorResponse(404, "专辑不存在"))
		return
	}
	sendJSON(w, 200, model.SuccessResponse(map[string]string{"message": "更新成功"}))
}

// DeleteAlbum DELETE /api/gallery/albums/:id
func DeleteAlbum(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		sendJSON(w, 405, model.ErrorResponse(405, "方法不允许"))
		return
	}
	userID := getGalleryUserID(r)
	id := extractID(r.URL.Path, "/api/gallery/albums/")

	// 将专辑内图片设为未分类
	db.GetDB().Model(&model.GalleryImage{}).
		Where("album_id = ? AND user_id = ?", id, userID).
		Update("album_id", nil)

	// 删除专辑
	result := db.GetDB().Where("id = ? AND user_id = ?", id, userID).
		Delete(&model.GalleryAlbum{})

	if result.RowsAffected == 0 {
		sendJSON(w, 404, model.ErrorResponse(404, "专辑不存在"))
		return
	}
	sendJSON(w, 200, model.SuccessResponse(map[string]string{"message": "删除成功"}))
}

// ============================================
// Day 3 新增：高级功能
// ============================================

// MoveImagesToAlbum POST /api/gallery/images/move - 批量移动图片到专辑
func MoveImagesToAlbum(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, 405, model.ErrorResponse(405, "方法不允许"))
		return
	}
	userID := getGalleryUserID(r)

	var req model.BatchMoveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, 400, model.ErrorResponse(400, "参数解析失败"))
		return
	}

	if len(req.IDs) == 0 {
		sendJSON(w, 400, model.ErrorResponse(400, "请选择要移动的图片"))
		return
	}

	// 验证目标album是否存在（如果指定了）
	if req.AlbumID != nil {
		var count int64
		db.GetDB().Model(&model.GalleryAlbum{}).Where("id = ? AND user_id = ?", *req.AlbumID, userID).Count(&count)
		if count == 0 {
			sendJSON(w, 404, model.ErrorResponse(404, "目标专辑不存在"))
			return
		}
	}

	// 执行批量更新
	result := db.GetDB().Model(&model.GalleryImage{}).
		Where("id IN ? AND user_id = ? AND is_deleted = ?", req.IDs, userID, false).
		Update("album_id", req.AlbumID)

	if result.Error != nil {
		sendJSON(w, 500, model.ErrorResponse(500, "移动失败"))
		return
	}

	sendJSON(w, 200, model.SuccessResponse(map[string]interface{}{
		"moved_count": result.RowsAffected,
		"message":     fmt.Sprintf("成功移动 %d 张图片", result.RowsAffected),
	}))
}

// GetRecycleBin GET /api/gallery/recycle-bin - 回收站列表
func GetRecycleBin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, 405, model.ErrorResponse(405, "方法不允许"))
		return
	}
	userID := getGalleryUserID(r)

	page, pageSize := 1, 50
	fmt.Sscanf(r.URL.Query().Get("page"), "%d", &page)
	fmt.Sscanf(r.URL.Query().Get("page_size"), "%d", &pageSize)
	if pageSize > 100 {
		pageSize = 100
	}

	// 查询30天内的已删除图片
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	query := db.GetDB().Model(&model.GalleryImage{}).
		Where("user_id = ? AND is_deleted = ? AND deleted_time > ?", userID, true, thirtyDaysAgo)

	var total int64
	query.Count(&total)

	var images []model.GalleryImage
	query.Order("deleted_time DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&images)

	items := make([]model.GalleryImageItem, len(images))
	for i, img := range images {
		items[i] = convertItem(img)
	}

	// 查询总共将删除多少图片（超过30天的）
	var willDeleteCount int64
	db.GetDB().Model(&model.GalleryImage{}).
		Where("user_id = ? AND is_deleted = ? AND deleted_time <= ?", userID, true, thirtyDaysAgo).
		Count(&willDeleteCount)

	sendJSON(w, 200, model.SuccessResponse(map[string]interface{}{
		"list":             items,
		"total":            total,
		"page":             page,
		"page_size":        pageSize,
		"total_pages":      int((total + int64(pageSize) - 1) / int64(pageSize)),
		"will_delete_soon": willDeleteCount,
		"expire_days":      30,
	}))
}

// RestoreImage PUT /api/gallery/recycle-bin/:id/restore - 恢复图片
func RestoreImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		sendJSON(w, 405, model.ErrorResponse(405, "方法不允许"))
		return
	}
	userID := getGalleryUserID(r)
	id := extractID(r.URL.Path, "/api/gallery/recycle-bin/")

	// 去掉 /restore 后缀
	id = strings.TrimSuffix(id, "/restore")

	result := db.GetDB().Model(&model.GalleryImage{}).
		Where("id = ? AND user_id = ? AND is_deleted = ?", id, userID, true).
		Updates(map[string]interface{}{
			"is_deleted":   false,
			"deleted_time": nil,
		})

	if result.RowsAffected == 0 {
		sendJSON(w, 404, model.ErrorResponse(404, "图片不存在或已过期"))
		return
	}
	sendJSON(w, 200, model.SuccessResponse(map[string]string{"message": "已恢复"}))
}

// PermanentDelete DELETE /api/gallery/recycle-bin/:id - 永久删除
func PermanentDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		sendJSON(w, 405, model.ErrorResponse(405, "方法不允许"))
		return
	}
	userID := getGalleryUserID(r)
	id := extractID(r.URL.Path, "/api/gallery/recycle-bin/")

	// 查询图片信息
	var img model.GalleryImage
	if err := db.GetDB().Where("id = ? AND user_id = ?", id, userID).First(&img).Error; err != nil {
		sendJSON(w, 404, model.ErrorResponse(404, "图片不存在"))
		return
	}

	// 删除物理文件
	os.Remove(img.FilePath)
	// 尝试删除缩略图（如果存在）
	thumbPath := filepath.Join(filepath.Dir(img.FilePath), "thumb_"+filepath.Base(img.FilePath))
	os.Remove(thumbPath)

	// 删除数据库记录
	db.GetDB().Unscoped().Delete(&img)

	sendJSON(w, 200, model.SuccessResponse(map[string]string{"message": "已永久删除"}))
}

// ============================================
// 辅助函数
// ============================================

func convertItem(img model.GalleryImage) model.GalleryImageItem {
	dir := filepath.Dir(img.FilePath)
	name := filepath.Base(img.FilePath)
	return model.GalleryImageItem{
		ID: img.ID, FilePath: img.FilePath,
		ThumbPath: filepath.Join(dir, "thumb_"+name),
		FileName:  img.FileName, FileSize: img.FileSize,
		UploadDate: img.UploadDate, Description: img.Description,
		Tags: img.Tags, IsFavorite: img.IsFavorite,
		CreatedAt: img.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func extractID(path, prefix string) string {
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	return strings.TrimPrefix(path, prefix)
}

// generateID 生成唯一ID（替换uuid）
func generateID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}
