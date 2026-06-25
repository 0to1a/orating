package humax

import "github.com/danielgtaylor/huma/v2"

func NotFound(msg string) error      { return huma.Error404NotFound(msg, nil) }
func Conflict(msg string) error      { return huma.Error409Conflict(msg, nil) }
func Forbidden(msg string) error     { return huma.Error403Forbidden(msg, nil) }
func Unprocessable(msg string) error { return huma.Error422UnprocessableEntity(msg, nil) }
func BadRequest(msg string) error    { return huma.Error400BadRequest(msg, nil) }
func Unauthorized(msg string) error  { return huma.Error401Unauthorized(msg, nil) }
