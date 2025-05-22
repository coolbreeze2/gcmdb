package client

import (
	"fmt"
	"strconv"
)

type MapKeyPathError struct {
	KeyPath string
}

func (e MapKeyPathError) Error() string {
	return fmt.Sprintf("map path %s doesn't exist", e.KeyPath)
}

type ResourceTypeError struct {
	Kind string
}

func (o ResourceTypeError) Error() string {
	return fmt.Sprintf("the server doesn't have a resource type %s", o.Kind)
}

type ResourceNotFoundError struct {
	Path      string
	Kind      string
	Name      string
	Namespace string
}

func (o ResourceNotFoundError) Error() string {
	msg := fmt.Sprintf("%s/%s not found at %s", o.Kind, o.Name, o.Path)
	if o.Namespace != "" {
		msg = fmt.Sprintf("%s/%s", o.Namespace, msg)
	}
	return msg
}

type ResourceValidateError struct {
	Path      string
	Kind      string
	Name      string
	Namespace string
	Message   string
}

func (o ResourceValidateError) Error() string {
	msg := fmt.Sprintf("%s/%s validate error %s at %s", o.Kind, o.Name, o.Message, o.Path)
	if o.Namespace != "" {
		msg = fmt.Sprintf("%s/%s", o.Namespace, msg)
	}
	return msg
}

type ResourceAlreadyExistError struct {
	Path      string
	Kind      string
	Name      string
	Namespace string
	Message   string
}

func (o ResourceAlreadyExistError) Error() string {
	msg := fmt.Sprintf("%s/%s already exist error %s at %s", o.Kind, o.Name, o.Message, o.Path)
	if o.Namespace != "" {
		msg = fmt.Sprintf("%s/%s", o.Namespace, msg)
	}
	return msg
}

type ResourceReferencedError struct {
	Path      string
	Kind      string
	Name      string
	Namespace string
	Message   string
}

func (o ResourceReferencedError) Error() string {
	msg := fmt.Sprintf("%s/%s has been referenced error %s at %s", o.Kind, o.Name, o.Message, o.Path)
	if o.Namespace != "" {
		msg = fmt.Sprintf("%s/%s", o.Namespace, msg)
	}
	return msg
}

type ServerError struct {
	Path       string
	StatusCode int
	Message    string
}

func (o ServerError) Error() string {
	msg := fmt.Sprintf("server response code %s Error at %s, %s", strconv.Itoa(o.StatusCode), o.Path, o.Message)
	return msg
}
