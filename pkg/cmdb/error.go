package cmdb

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
	Message   string
}

func (o ResourceNotFoundError) Error() string {
	return fmtNamespaceError(
		fmt.Sprintf("%s/%s not found at %s %s", o.Kind, o.Name, o.Path, o.Message),
		o.Namespace,
	)
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
	return fmtNamespaceError(msg, o.Namespace)
}

type ResourceAlreadyExistError struct {
	Path      string
	Kind      string
	Name      string
	Namespace string
	Message   string
}

func (o ResourceAlreadyExistError) Error() string {
	return fmtNamespaceError(
		fmt.Sprintf("%s/%s already exist error %s at %s", o.Kind, o.Name, o.Message, o.Path),
		o.Namespace,
	)
}

type ResourceReferencedError struct {
	Path      string
	Kind      string
	Name      string
	Namespace string
	Message   string
}

func (o ResourceReferencedError) Error() string {
	return fmtNamespaceError(
		fmt.Sprintf("%s/%s has been referenced error %s at %s", o.Kind, o.Name, o.Message, o.Path),
		o.Namespace,
	)
}

func fmtNamespaceError(msg, namespace string) string {
	if namespace != "" {
		msg = fmt.Sprintf("%s/%s", namespace, msg)
	}
	return msg
}

type ServerError struct {
	Path       string
	StatusCode int
	Message    string
}

func (o ServerError) Error() string {
	return fmt.Sprintf("server response code %s Error at %s, %s", strconv.Itoa(o.StatusCode), o.Path, o.Message)
}
