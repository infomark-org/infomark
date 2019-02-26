// .............................................................................
package app

import (
  "net/http"

  validation "github.com/go-ozzo/ozzo-validation"
)

type H map[string]interface{}

// userRequest is the request payload for user management.
type EmailRequest struct {
  Subject string `json:"subject"`
  Body    string `json:"body"`
}

// Bind preprocesses a userRequest.
func (body *EmailRequest) Bind(r *http.Request) error {

  err := validation.ValidateStruct(body,
    validation.Field(&body.Body, validation.Required),
    validation.Field(&body.Subject, validation.Required),
  )

  return err
}
