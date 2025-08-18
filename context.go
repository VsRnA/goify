package goify

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

type Context struct {
	Request  *http.Request
	Response http.ResponseWriter
	params   map[string]string
	store    map[string]interface{}
}

func (c *Context) Param(key string) string {
	return c.params[key]
}

func (c *Context) Query(key string) string {
	return c.Request.URL.Query().Get(key)
}

func (c *Context) QueryDefault(key, defaultValue string) string {
	value := c.Query(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func (c *Context) QueryInt(key string) (int, error) {
	value := c.Query(key)
	if value == "" {
		return 0, fmt.Errorf("query parameter '%s' not found", key)
	}
	return strconv.Atoi(value)
}

func (c *Context) Body() ([]byte, error) {
	defer c.Request.Body.Close()
	buf := make([]byte, c.Request.ContentLength)
	_, err := c.Request.Body.Read(buf)
	return buf, err
}

func (c *Context) BindJSON(obj interface{}) error {
	decoder := json.NewDecoder(c.Request.Body)
	return decoder.Decode(obj)
}

func (c *Context) BindAndValidate(obj interface{}) error {
	if err := c.BindJSON(obj); err != nil {
		return err
	}
	
	if validationErrors := Validate(obj); len(validationErrors) > 0 {
		return validationErrors
	}
	
	return nil
}

func (c *Context) ValidateStruct(obj interface{}) ValidationErrors {
	return Validate(obj)
}

func (c *Context) ValidateQuery(obj interface{}) error {
	rv := reflect.ValueOf(obj)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("obj must be a pointer to struct")
	}
	
	rv = rv.Elem()
	rt := rv.Type()
	
	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		fieldType := rt.Field(i)
		
		if !field.CanSet() {
			continue
		}

		paramName := fieldType.Name
		if tag := fieldType.Tag.Get("query"); tag != "" {
			paramName = tag
		} else if tag := fieldType.Tag.Get("json"); tag != "" {
			if tagName := strings.Split(tag, ",")[0]; tagName != "" && tagName != "-" {
				paramName = tagName
			}
		}
		
		queryValue := c.Query(paramName)
		if queryValue == "" {
			continue
		}

		if err := setFieldValue(field, queryValue); err != nil {
			return fmt.Errorf("invalid value for field %s: %v", fieldType.Name, err)
		}
	}

	if validationErrors := Validate(obj); len(validationErrors) > 0 {
		return validationErrors
	}
	
	return nil
}

func setFieldValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(intVal)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(uintVal)
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(floatVal)
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(boolVal)
	default:
		return fmt.Errorf("unsupported field type: %v", field.Kind())
	}
	return nil
}

func (c *Context) GetHeader(key string) string {
	return c.Request.Header.Get(key)
}

func (c *Context) SetHeader(key, value string) {
	c.Response.Header().Set(key, value)
}

func (c *Context) Status(code int) *Context {
	c.Response.WriteHeader(code)
	return c
}

func (c *Context) JSON(code int, obj interface{}) error {
	c.SetHeader("Content-Type", "application/json")
	c.Response.WriteHeader(code)
	return json.NewEncoder(c.Response).Encode(obj)
}

func (c *Context) String(code int, format string, values ...interface{}) error {
	c.SetHeader("Content-Type", "text/plain")
	c.Response.WriteHeader(code)
	_, err := fmt.Fprintf(c.Response, format, values...)
	return err
}

func (c *Context) HTML(code int, html string) error {
	c.SetHeader("Content-Type", "text/html")
	c.Response.WriteHeader(code)
	_, err := c.Response.Write([]byte(html))
	return err
}

func (c *Context) Redirect(code int, location string) error {
	if code < 300 || code > 308 {
		return fmt.Errorf("invalid redirect status code: %d", code)
	}
	c.SetHeader("Location", location)
	c.Response.WriteHeader(code)
	return nil
}

func (c *Context) Cookie(name string) (*http.Cookie, error) {
	return c.Request.Cookie(name)
}

func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Response, cookie)
}

func (c *Context) Form(key string) string {
	return c.Request.FormValue(key)
}

func (c *Context) FormFile(key string) (*FileHeader, error) {
	file, header, err := c.Request.FormFile(key)
	if err != nil {
		return nil, err
	}
	
	return &FileHeader{
		FileHeader: header,
		File:       file,
	}, nil
}

func (c *Context) FormFiles(key string) ([]*FileHeader, error) {
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		return nil, err
	}
	
	if c.Request.MultipartForm == nil || c.Request.MultipartForm.File == nil {
		return nil, fmt.Errorf("no multipart form")
	}
	
	fileHeaders := c.Request.MultipartForm.File[key]
	if len(fileHeaders) == 0 {
		return nil, fmt.Errorf("no files found for key: %s", key)
	}
	
	var files []*FileHeader
	for _, fh := range fileHeaders {
		file, err := fh.Open()
		if err != nil {
			return nil, err
		}
		
		files = append(files, &FileHeader{
			FileHeader: fh,
			File:       file,
		})
	}
	
	return files, nil
}

func (c *Context) SaveUploadedFile(fileHeader *FileHeader, uploadDir string) (string, error) {
	if fileHeader == nil {
		return "", fmt.Errorf("file header is nil")
	}

	filename := GenerateUniqueFilename(fileHeader.Filename)

	err := SaveFileWithName(fileHeader, uploadDir, filename)
	if err != nil {
		return "", err
	}
	
	return filepath.Join(uploadDir, filename), nil
}

func (c *Context) ValidateFile(fileHeader *FileHeader, validation FileValidation) error {
	return ValidateFile(fileHeader, validation)
}

func (c *Context) ValidateFiles(fileHeaders []*FileHeader, validation FileValidation) FileUploadErrors {
	var errors FileUploadErrors
	
	for i, fileHeader := range fileHeaders {
		if err := ValidateFile(fileHeader, validation); err != nil {
			if uploadErr, ok := err.(FileUploadError); ok {
				uploadErr.Field = fmt.Sprintf("file[%d]", i)
				errors = append(errors, uploadErr)
			}
		}
	}
	
	return errors
}

func (c *Context) BindMultipart(obj interface{}) error {
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		return err
	}
	
	rv := reflect.ValueOf(obj)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("obj must be a pointer to struct")
	}
	
	rv = rv.Elem()
	rt := rv.Type()
	
	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		fieldType := rt.Field(i)
		
		if !field.CanSet() {
			continue
		}

		fieldName := fieldType.Name
		if tag := fieldType.Tag.Get("form"); tag != "" {
			fieldName = tag
		}

		if field.Type() == reflect.TypeOf(&FileHeader{}) {
			fileHeader, err := c.FormFile(fieldName)
			if err == nil {
				field.Set(reflect.ValueOf(fileHeader))
			}
			continue
		}

		if field.Type() == reflect.TypeOf([]*FileHeader{}) {
			fileHeaders, err := c.FormFiles(fieldName)
			if err == nil {
				field.Set(reflect.ValueOf(fileHeaders))
			}
			continue
		}

		formValue := c.Request.FormValue(fieldName)
		if formValue == "" {
			continue
		}
		
		if err := setFieldValue(field, formValue); err != nil {
			return fmt.Errorf("invalid value for field %s: %v", fieldType.Name, err)
		}
	}
	
	return nil
}

func (c *Context) GetUploadedFileInfo(key string) (map[string]interface{}, error) {
	fileHeader, err := c.FormFile(key)
	if err != nil {
		return nil, err
	}
	
	return map[string]interface{}{
		"filename":    fileHeader.Filename,
		"size":        fileHeader.Size,
		"size_human":  FormatFileSize(fileHeader.Size),
		"content_type": fileHeader.Header.Get("Content-Type"),
		"is_image":    IsImageFile(fileHeader.Header.Get("Content-Type")),
	}, nil
}

func (c *Context) setParam(key, value string) {
	if c.params == nil {
		c.params = make(map[string]string)
	}
	decoded, _ := url.QueryUnescape(value)
	c.params[key] = decoded
}

func (c *Context) Set(key string, value interface{}) {
	if c.store == nil {
		c.store = make(map[string]interface{})
	}
	c.store[key] = value
}

func (c *Context) Get(key string) (interface{}, bool) {
	if c.store == nil {
		return nil, false
	}
	value, exists := c.store[key]
	return value, exists
}

func (c *Context) MustGet(key string) interface{} {
	if value, exists := c.Get(key); exists {
		return value
	}
	panic("Key \"" + key + "\" does not exist")
}