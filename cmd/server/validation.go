package main

import (
	"strconv"
	"strings"
)

// Validator accumulates validation errors for a single request.
type Validator struct {
	errors []string
}

// NewValidator creates a new Validator with no errors.
func NewValidator() *Validator {
	return &Validator{}
}

// Required checks that value is non-empty.
func (v *Validator) Required(field, value string) *Validator {
	if strings.TrimSpace(value) == "" {
		v.errors = append(v.errors, field+" é obrigatório")
	}
	return v
}

// MinLength checks minimum string length.
func (v *Validator) MinLength(field, value string, min int) *Validator {
	if len(value) < min {
		v.errors = append(v.errors, field+" deve ter pelo menos "+strconv.Itoa(min)+" caracteres")
	}
	return v
}

// ValidRole checks that role is one of the allowed values.
func (v *Validator) ValidRole(field, role string) *Validator {
	if !validRole(role) {
		v.errors = append(v.errors, field+" inválido")
	}
	return v
}

// Positive checks that value > 0.
func (v *Validator) Positive(field string, value int) *Validator {
	if value <= 0 {
		v.errors = append(v.errors, field+" deve ser positivo")
	}
	return v
}

// IsValid returns true if no errors accumulated.
func (v *Validator) IsValid() bool {
	return len(v.errors) == 0
}

// Errors returns all accumulated error messages.
func (v *Validator) Errors() []string {
	return v.errors
}

// Error returns a single combined message (for AppError wrapping).
func (v *Validator) Error() string {
	return strings.Join(v.errors, "; ")
}
