package client

import (
	"fmt"
	"strconv"
)

type ListOptions struct {
	Namespace     string            `json:"namespace"`
	Page          int64             `json:"page"`
	Limit         int64             `json:"limit"`
	Selector      map[string]string `json:"selector"`
	FieldSelector map[string]string `json:"field_selector"`
}

type HttpRequestArgs struct {
	Method  string
	Url     string
	Query   map[string]string
	Headers map[string]string
	Data    any
}

type ObjectNotFoundError struct {
	path      string
	kind      string
	name      string
	namespace string
}

func (o ObjectNotFoundError) Error() string {
	msg := fmt.Sprintf("%s/%s not found at %s", o.kind, o.name, o.path)
	if o.namespace != "" {
		msg = fmt.Sprintf("%s/%s", o.namespace, msg)
	}
	return msg
}

type ObjectValidateError struct {
	path      string
	kind      string
	name      string
	namespace string
	message   string
}

func (o ObjectValidateError) Error() string {
	msg := fmt.Sprintf("%s/%s validate error %s at %s", o.kind, o.name, o.message, o.path)
	if o.namespace != "" {
		msg = fmt.Sprintf("%s/%s", o.namespace, msg)
	}
	return msg
}

type ObjectAlreadyExistError struct {
	path      string
	kind      string
	name      string
	namespace string
	message   string
}

func (o ObjectAlreadyExistError) Error() string {
	msg := fmt.Sprintf("%s/%s already exist error %s at %s", o.kind, o.name, o.message, o.path)
	if o.namespace != "" {
		msg = fmt.Sprintf("%s/%s", o.namespace, msg)
	}
	return msg
}

type ObjectReferencedError struct {
	path      string
	kind      string
	name      string
	namespace string
	message   string
}

func (o ObjectReferencedError) Error() string {
	msg := fmt.Sprintf("%s/%s has been referenced error %s at %s", o.kind, o.name, o.message, o.path)
	if o.namespace != "" {
		msg = fmt.Sprintf("%s/%s", o.namespace, msg)
	}
	return msg
}

type ServerError struct {
	path       string
	statusCode int
	message    string
}

func (o ServerError) Error() string {
	msg := fmt.Sprintf("Server response code %s Error at %s, %s", strconv.Itoa(o.statusCode), o.path, o.message)
	return msg
}
