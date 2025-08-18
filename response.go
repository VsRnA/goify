package goify

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type H map[string]interface{}

type ErrorResponse struct {
	Error   string      `json:"error"`
	Message string      `json:"message,omitempty"`
	Code    int         `json:"code"`
	Details interface{} `json:"details,omitempty"`
}

type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

func (c *Context) SendError(code int, message string, details ...interface{}) error {
	errorResp := ErrorResponse{
		Error:   http.StatusText(code),
		Message: message,
		Code:    code,
	}
	
	if len(details) > 0 {
		errorResp.Details = details[0]
	}
	
	return c.JSON(code, errorResp)
}

func (c *Context) SendSuccess(data interface{}, message ...string) error {
	successResp := SuccessResponse{
		Success: true,
		Data:    data,
	}
	
	if len(message) > 0 {
		successResp.Message = message[0]
	}
	
	return c.JSON(http.StatusOK, successResp)
}

func (c *Context) SendCreated(data interface{}, message ...string) error {
	successResp := SuccessResponse{
		Success: true,
		Data:    data,
	}
	
	if len(message) > 0 {
		successResp.Message = message[0]
	}
	
	return c.JSON(http.StatusCreated, successResp)
}

func (c *Context) SendNoContent() error {
	c.Response.WriteHeader(http.StatusNoContent)
	return nil
}

func (c *Context) SendNotFound(message ...string) error {
	msg := "Resource not found"
	if len(message) > 0 {
		msg = message[0]
	}
	return c.SendError(http.StatusNotFound, msg)
}

func (c *Context) SendBadRequest(message string, details ...interface{}) error {
	return c.SendError(http.StatusBadRequest, message, details...)
}

func (c *Context) SendValidationError(validationErrors interface{}) error {
	var message string
	var details interface{}
	
	switch ve := validationErrors.(type) {
	case ValidationErrors:
		message = "Validation failed"
		details = ve
	case error:
		message = ve.Error()
		if validationErrs, ok := validationErrors.(ValidationErrors); ok {
			details = validationErrs
		}
	default:
		message = "Validation failed"
		details = validationErrors
	}
	
	errorResp := ErrorResponse{
		Error:   "Validation Error",
		Message: message,
		Code:    422,
		Details: details,
	}
	
	return c.JSON(422, errorResp)
}

func (c *Context) SendFieldError(field, message string) error {
	validationError := ValidationErrors{
		{
			Field:   field,
			Message: message,
		},
	}
	return c.SendValidationError(validationError)
}

func (c *Context) SendFileUploadError(uploadErrors interface{}) error {
	var message string
	var details interface{}
	
	switch ue := uploadErrors.(type) {
	case FileUploadErrors:
		message = "File upload failed"
		details = ue
	case FileUploadError:
		message = ue.Message
		details = []FileUploadError{ue}
	case error:
		message = ue.Error()
	default:
		message = "File upload failed"
		details = uploadErrors
	}
	
	errorResp := ErrorResponse{
		Error:   "File Upload Error",
		Message: message,
		Code:    422,
		Details: details,
	}
	
	return c.JSON(422, errorResp)
}

func (c *Context) SendFileTooBigError(maxSize int64) error {
	return c.SendFileUploadError(FileUploadError{
		Message: fmt.Sprintf("File size exceeds maximum allowed size of %s", FormatFileSize(maxSize)),
		Code:    "file_too_big",
	})
}

func (c *Context) SendUnauthorized(message ...string) error {
	msg := "Unauthorized access"
	if len(message) > 0 {
		msg = message[0]
	}
	return c.SendError(http.StatusUnauthorized, msg)
}

func (c *Context) SendForbidden(message ...string) error {
	msg := "Access forbidden"
	if len(message) > 0 {
		msg = message[0]
	}
	return c.SendError(http.StatusForbidden, msg)
}

func (c *Context) SendInternalError(message ...string) error {
	msg := "Internal server error"
	if len(message) > 0 {
		msg = message[0]
	}
	return c.SendError(http.StatusInternalServerError, msg)
}

func (c *Context) SendFile(filepath string) error {
	http.ServeFile(c.Response, c.Request, filepath)
	return nil
}

func (c *Context) Download(filepath, filename string) error {
	if filename != "" {
		c.SetHeader("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	}
	http.ServeFile(c.Response, c.Request, filepath)
	return nil
}

func (c *Context) Stream(contentType string, fn func(http.ResponseWriter)) error {
	c.SetHeader("Content-Type", contentType)
	c.SetHeader("Transfer-Encoding", "chunked")
	fn(c.Response)
	return nil
}

func (c *Context) JSONPretty(code int, obj interface{}, indent string) error {
	c.SetHeader("Content-Type", "application/json")
	c.Response.WriteHeader(code)
	
	encoder := json.NewEncoder(c.Response)
	encoder.SetIndent("", indent)
	return encoder.Encode(obj)
}