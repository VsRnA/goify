package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/VsRnA/goify"
)

type FileUploadRequest struct {
	Title       string              `form:"title" validate:"required,min=3,max=100"`
	Description string              `form:"description" validate:"max=500"`
	Category    string              `form:"category" validate:"required,oneof=image document video"`
	File        *goify.FileHeader   `form:"file"`
	Files       []*goify.FileHeader `form:"files"`
	IsPublic    bool                `form:"is_public"`
}

func main() {
	app := goify.New()

	app.Use(goify.Logger())
	app.Use(goify.Recovery())
	app.Use(goify.CORS())

	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatal("Failed to create uploads directory:", err)
	}

	app.Use(goify.Static("/uploads", "./uploads"))

	app.GET("/", func(c *goify.Context) {
		c.HTML(200, `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Goify File Upload Example</title>
			<style>
				body { font-family: Arial, sans-serif; margin: 40px; }
				.form-group { margin-bottom: 15px; }
				label { display: block; margin-bottom: 5px; font-weight: bold; }
				input, select, textarea { width: 300px; padding: 8px; }
				button { background: #007bff; color: white; padding: 10px 20px; border: none; cursor: pointer; }
				.result { margin-top: 20px; padding: 10px; background: #f8f9fa; border: 1px solid #dee2e6; }
			</style>
		</head>
		<body>
			<h1>Goify File Upload Example</h1>
			
			<h2>Single File Upload</h2>
			<form action="/upload/single" method="post" enctype="multipart/form-data">
				<div class="form-group">
					<label>Title:</label>
					<input type="text" name="title" required>
				</div>
				<div class="form-group">
					<label>Category:</label>
					<select name="category" required>
						<option value="image">Image</option>
						<option value="document">Document</option>
						<option value="video">Video</option>
					</select>
				</div>
				<div class="form-group">
					<label>File:</label>
					<input type="file" name="file" required>
				</div>
				<div class="form-group">
					<label>Description:</label>
					<textarea name="description" rows="3"></textarea>
				</div>
				<div class="form-group">
					<label>
						<input type="checkbox" name="is_public"> Make public
					</label>
				</div>
				<button type="submit">Upload Single File</button>
			</form>

			<h2>Multiple Files Upload</h2>
			<form action="/upload/multiple" method="post" enctype="multipart/form-data">
				<div class="form-group">
					<label>Title:</label>
					<input type="text" name="title" required>
				</div>
				<div class="form-group">
					<label>Files (multiple):</label>
					<input type="file" name="files" multiple required>
				</div>
				<button type="submit">Upload Multiple Files</button>
			</form>

			<h2>Avatar Upload (Images Only)</h2>
			<form action="/upload/avatar" method="post" enctype="multipart/form-data">
				<div class="form-group">
					<label>Avatar:</label>
					<input type="file" name="avatar" accept="image/*" required>
				</div>
				<button type="submit">Upload Avatar</button>
			</form>

			<h2>Endpoints:</h2>
			<ul>
				<li><code>POST /upload/single</code> - Single file upload with validation</li>
				<li><code>POST /upload/multiple</code> - Multiple files upload</li>
				<li><code>POST /upload/avatar</code> - Avatar upload (images only)</li>
				<li><code>POST /upload/stream</code> - Stream upload for large files</li>
				<li><code>GET /files</code> - List uploaded files</li>
				<li><code>GET /uploads/*</code> - Serve uploaded files</li>
			</ul>
		</body>
		</html>`)
	})

	app.POST("/upload/single", func(c *goify.Context) {
		var req FileUploadRequest

		if err := c.BindMultipart(&req); err != nil {
			c.SendBadRequest("Failed to parse form data", err.Error())
			return
		}

		if validationErrors := c.ValidateStruct(&req); len(validationErrors) > 0 {
			c.SendValidationError(validationErrors)
			return
		}

		if req.File == nil {
			c.SendFieldError("file", "File is required")
			return
		}

		var validation goify.FileValidation
		switch req.Category {
		case "image":
			validation = goify.FileValidation{
				MaxSize:      5 * 1024 * 1024,
				AllowedTypes: []string{"image/jpeg", "image/png", "image/gif"},
				AllowedExts:  []string{".jpg", ".jpeg", ".png", ".gif"},
				Required:     true,
			}
		case "document":
			validation = goify.FileValidation{
				MaxSize:      10 * 1024 * 1024,
				AllowedTypes: []string{"application/pdf", "text/plain", "application/msword"},
				AllowedExts:  []string{".pdf", ".txt", ".doc", ".docx"},
				Required:     true,
			}
		case "video":
			validation = goify.FileValidation{
				MaxSize:      50 * 1024 * 1024,
				AllowedTypes: []string{"video/mp4", "video/avi", "video/quicktime"},
				AllowedExts:  []string{".mp4", ".avi", ".mov"},
				Required:     true,
			}
		}

		if err := c.ValidateFile(req.File, validation); err != nil {
			c.SendFileUploadError(err)
			return
		}

		savedPath, err := c.SaveUploadedFile(req.File, uploadDir)
		if err != nil {
			c.SendInternalError("Failed to save file")
			return
		}

		fileInfo, _ := c.GetUploadedFileInfo("file")

		c.SendCreated(goify.H{
			"message":     "File uploaded successfully",
			"title":       req.Title,
			"description": req.Description,
			"category":    req.Category,
			"is_public":   req.IsPublic,
			"file": goify.H{
				"original_name": req.File.Filename,
				"saved_path":    savedPath,
				"size":          fileInfo["size"],
				"size_human":    fileInfo["size_human"],
				"content_type":  fileInfo["content_type"],
				"url":           "/uploads/" + filepath.Base(savedPath),
			},
		})
	})

	app.POST("/upload/multiple", func(c *goify.Context) {
		var req struct {
			Title string                `form:"title" validate:"required"`
			Files []*goify.FileHeader   `form:"files"`
		}

		if err := c.BindMultipart(&req); err != nil {
			c.SendBadRequest("Failed to parse form data")
			return
		}

		if validationErrors := c.ValidateStruct(&req); len(validationErrors) > 0 {
			c.SendValidationError(validationErrors)
			return
		}

		if len(req.Files) == 0 {
			c.SendFieldError("files", "At least one file is required")
			return
		}

		validation := goify.FileValidation{
			MaxSize:     5 * 1024 * 1024,
			Required:    true,
		}

		if uploadErrors := c.ValidateFiles(req.Files, validation); len(uploadErrors) > 0 {
			c.SendFileUploadError(uploadErrors)
			return
		}

		var savedFiles []goify.H
		for _, file := range req.Files {
			savedPath, err := c.SaveUploadedFile(file, uploadDir)
			if err != nil {
				c.SendInternalError("Failed to save one or more files")
				return
			}

			savedFiles = append(savedFiles, goify.H{
				"original_name": file.Filename,
				"saved_path":    savedPath,
				"size":          file.Size,
				"size_human":    goify.FormatFileSize(file.Size),
				"content_type":  file.Header.Get("Content-Type"),
				"url":           "/uploads/" + filepath.Base(savedPath),
			})
		}

		c.SendCreated(goify.H{
			"message": "Files uploaded successfully",
			"title":   req.Title,
			"files":   savedFiles,
			"count":   len(savedFiles),
		})
	})

	app.POST("/upload/avatar", func(c *goify.Context) {
		file, err := c.FormFile("avatar")
		if err != nil {
			c.SendBadRequest("Avatar file is required")
			return
		}

		validation := goify.FileValidation{
			MaxSize:      2 * 1024 * 1024,
			AllowedTypes: []string{"image/jpeg", "image/png"},
			AllowedExts:  []string{".jpg", ".jpeg", ".png"},
			Required:     true,
		}

		if err := c.ValidateFile(file, validation); err != nil {
			c.SendFileUploadError(err)
			return
		}

		avatarDir := filepath.Join(uploadDir, "avatars")
		savedPath, err := c.SaveUploadedFile(file, avatarDir)
		if err != nil {
			c.SendInternalError("Failed to save avatar")
			return
		}

		c.SendCreated(goify.H{
			"message": "Avatar uploaded successfully",
			"avatar": goify.H{
				"original_name": file.Filename,
				"saved_path":    savedPath,
				"size_human":    goify.FormatFileSize(file.Size),
				"url":           "/uploads/avatars/" + filepath.Base(savedPath),
			},
		})
	})

	app.POST("/upload/stream", func(c *goify.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.SendBadRequest("File is required")
			return
		}

		filename := goify.GenerateUniqueFilename(file.Filename)
		dst := filepath.Join(uploadDir, "large", filename)

		if err := goify.SaveFileWithName(file, filepath.Dir(dst), filepath.Base(dst)); err != nil {
			c.SendInternalError("Failed to save large file")
			return
		}

		c.SendCreated(goify.H{
			"message":  "Large file uploaded successfully",
			"filename": filename,
			"path":     dst,
			"size":     goify.FormatFileSize(file.Size),
		})
	})

	app.GET("/files", func(c *goify.Context) {
		files := []goify.H{}

		err := filepath.Walk(uploadDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				relPath, _ := filepath.Rel(uploadDir, path)
				files = append(files, goify.H{
					"name":       info.Name(),
					"path":       relPath,
					"size":       info.Size(),
					"size_human": goify.FormatFileSize(info.Size()),
					"modified":   info.ModTime(),
					"url":        "/uploads/" + relPath,
				})
			}
			return nil
		})

		if err != nil {
			c.SendInternalError("Failed to list files")
			return
		}

		c.SendSuccess(goify.H{
			"files": files,
			"count": len(files),
		})
	})

	app.GET("/files/:filename", func(c *goify.Context) {
		filename := c.Param("filename")
		filePath := filepath.Join(uploadDir, filename)

		if !goify.FileExists(filePath) {
			c.SendNotFound("File not found")
			return
		}

		size, _ := goify.GetFileSize(filePath)
		mimeType := goify.GetMimeType(filename)

		c.SendSuccess(goify.H{
			"filename":     filename,
			"size":         size,
			"size_human":   goify.FormatFileSize(size),
			"content_type": mimeType,
			"is_image":     goify.IsImageFile(mimeType),
			"url":          "/uploads/" + filename,
		})
	})

	app.DELETE("/files/:filename", func(c *goify.Context) {
		filename := c.Param("filename")
		filePath := filepath.Join(uploadDir, filename)

		if !goify.FileExists(filePath) {
			c.SendNotFound("File not found")
			return
		}

		if err := goify.DeleteFile(filePath); err != nil {
			c.SendInternalError("Failed to delete file")
			return
		}

		c.SendSuccess(goify.H{
			"message":  "File deleted successfully",
			"filename": filename,
		})
	})

	app.POST("/api/upload", func(c *goify.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.SendBadRequest("File is required")
			return
		}

		title := c.Request.FormValue("title")
		if title == "" {
			c.SendFieldError("title", "Title is required")
			return
		}

		validation := goify.FileValidation{
			MaxSize:     10 * 1024 * 1024,
			Required:    true,
		}

		if err := c.ValidateFile(file, validation); err != nil {
			c.SendFileUploadError(err)
			return
		}

		savedPath, err := c.SaveUploadedFile(file, uploadDir)
		if err != nil {
			c.SendInternalError("Failed to save file")
			return
		}

		c.JSON(201, goify.H{
			"success": true,
			"data": goify.H{
				"id":           12345,
				"title":        title,
				"filename":     file.Filename,
				"saved_path":   savedPath,
				"size":         file.Size,
				"size_human":   goify.FormatFileSize(file.Size),
				"content_type": file.Header.Get("Content-Type"),
				"url":          "/uploads/" + filepath.Base(savedPath),
				"uploaded_at":  "2024-01-01T12:00:00Z",
			},
		})
	})

	log.Println("Server started with File Upload support!")
	log.Println("")
	log.Println("Available endpoints:")
	log.Println("  GET  / - Upload form interface")
	log.Println("  POST /upload/single - Single file upload with validation")
	log.Println("  POST /upload/multiple - Multiple files upload")
	log.Println("  POST /upload/avatar - Avatar upload (images only)")
	log.Println("  POST /upload/stream - Stream upload for large files")
	log.Println("  POST /api/upload - JSON API file upload")
	log.Println("  GET  /files - List all uploaded files")
	log.Println("  GET  /files/:filename - Get file information")
	log.Println("  DELETE /files/:filename - Delete a file")
	log.Println("  GET  /uploads/* - Serve uploaded files")
	log.Println("")
	log.Println("Upload directory:", uploadDir)
	log.Println("")
	log.Println("Try uploading files:")
	log.Println("  Open http://localhost:3000 in your browser")
	log.Println("  Or use curl:")
	log.Println(`    curl -X POST -F "title=Test" -F "category=image" -F "file=@image.jpg" http://localhost:3000/upload/single`)

	if err := app.Listen(":3000"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}