package client

import (
	"fmt"
	"strconv"
)

type MapKeyPathError struct {
	keyPath string
}

func (e MapKeyPathError) Error() string {
	return fmt.Sprintf("Map path %s doesn't exist", e.keyPath)
}

type ResourceNotFoundError struct {
	path      string
	kind      string
	name      string
	namespace string
}

func (o ResourceNotFoundError) Error() string {
	msg := fmt.Sprintf("%s/%s not found at %s", o.kind, o.name, o.path)
	if o.namespace != "" {
		msg = fmt.Sprintf("%s/%s", o.namespace, msg)
	}
	return msg
}

type ResourceValidateError struct {
	path      string
	kind      string
	name      string
	namespace string
	message   string
}

func (o ResourceValidateError) Error() string {
	msg := fmt.Sprintf("%s/%s validate error %s at %s", o.kind, o.name, o.message, o.path)
	if o.namespace != "" {
		msg = fmt.Sprintf("%s/%s", o.namespace, msg)
	}
	return msg
}

type ResourceAlreadyExistError struct {
	path      string
	kind      string
	name      string
	namespace string
	message   string
}

func (o ResourceAlreadyExistError) Error() string {
	msg := fmt.Sprintf("%s/%s already exist error %s at %s", o.kind, o.name, o.message, o.path)
	if o.namespace != "" {
		msg = fmt.Sprintf("%s/%s", o.namespace, msg)
	}
	return msg
}

type ResourceReferencedError struct {
	path      string
	kind      string
	name      string
	namespace string
	message   string
}

func (o ResourceReferencedError) Error() string {
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
