package openapi3

import (
	"context"
	"errors"
	"fmt"

	"github.com/getkin/kin-openapi/jsoninfo"
)

// T is the root of an OpenAPI v3 document
// See https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.3.md#openapi-object
type T struct {
	ExtensionProps `json:"-" yaml:"-"`

	OpenAPI      Version              `json:"openapi" yaml:"openapi"` // Required
	Components   Components           `json:"components,omitempty" yaml:"components,omitempty"`
	Info         *Info                `json:"info" yaml:"info"`   // Required
	Paths        Paths                `json:"paths" yaml:"paths"` // Required
	Security     SecurityRequirements `json:"security,omitempty" yaml:"security,omitempty"`
	Servers      Servers              `json:"servers,omitempty" yaml:"servers,omitempty"`
	Tags         Tags                 `json:"tags,omitempty" yaml:"tags,omitempty"`
	ExternalDocs *ExternalDocs        `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`

	visited          visitedComponent
	specMinorVersion uint64 // defaults to 0 (3.0.z)
}

// MarshalJSON returns the JSON encoding of T.
func (doc *T) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalStrictStruct(doc)
}

// UnmarshalJSON sets T to a copy of data.
func (doc *T) UnmarshalJSON(data []byte) error {
	err := jsoninfo.UnmarshalStrictStruct(data, doc)
	if err != nil {
		doc.WithMinorOpenAPIVersion(doc.OpenAPI.Minor())
	}
	return err
}

// WithMinorOpenAPIVersion allows to enable specification minor feature version
func (doc *T) WithMinorOpenAPIVersion(minorVersion uint64) *T {
	if doc != nil {
		// store so it can be used in validation context if needed
		doc.specMinorVersion = minorVersion
		// propagate down to internal properties
		doc.Components.WithMinorOpenAPIVersion(minorVersion)
		doc.Info.WithMinorOpenAPIVersion(minorVersion)
		doc.Paths.WithMinorOpenAPIVersion(minorVersion)
		doc.Security.WithMinorOpenAPIVersion(minorVersion)
		doc.Servers.WithMinorOpenAPIVersion(minorVersion)
		doc.Tags.WithMinorOpenAPIVersion(minorVersion)
		doc.ExternalDocs.WithMinorOpenAPIVersion(minorVersion)
	}
	return doc
}

// AddOperation updates T wih a new OpenAPI operation for given method and path
func (doc *T) AddOperation(path string, method string, operation *Operation) {
	paths := doc.Paths
	if paths == nil {
		paths = make(Paths)
		doc.Paths = paths
	}
	pathItem := paths[path]
	if pathItem == nil {
		pathItem = &PathItem{}
		paths[path] = pathItem
	}
	pathItem.SetOperation(method, operation)
	operation.WithMinorOpenAPIVersion(doc.specMinorVersion)
}

func (doc *T) AddServer(server *Server) {
	doc.Servers = append(doc.Servers, server)
}

// Validate returns an error if T does not comply with the OpenAPI spec.
// Validations Options can be provided to modify the validation behavior.
func (doc *T) Validate(ctx context.Context, opts ...ValidationOption) error {
	ctx = WithValidationOptions(ctx, opts...)

	var wrap func(error) error
	// NOTE: only mention info/components/paths/... key in this func's errors.

	wrap = func(e error) error { return fmt.Errorf("invalid openapi value: %w", e) }
	if doc.OpenAPI == "" {
		return wrap(errors.New("must be a non-empty string"))
	}
	if err := doc.OpenAPI.Validate(ctx); err != nil {
		return wrap(err)
	}

	wrap = func(e error) error { return fmt.Errorf("invalid components: %w", e) }
	if err := doc.Components.Validate(ctx); err != nil {
		return wrap(err)
	}

	wrap = func(e error) error { return fmt.Errorf("invalid info: %w", e) }
	if v := doc.Info; v != nil {
		if err := v.Validate(ctx); err != nil {
			return wrap(err)
		}
	} else {
		return wrap(errors.New("must be an object"))
	}

	wrap = func(e error) error { return fmt.Errorf("invalid paths: %w", e) }
	if v := doc.Paths; v != nil {
		if err := v.Validate(ctx); err != nil {
			return wrap(err)
		}
	} else {
		return wrap(errors.New("must be an object"))
	}

	wrap = func(e error) error { return fmt.Errorf("invalid security: %w", e) }
	if v := doc.Security; v != nil {
		if err := v.Validate(ctx); err != nil {
			return wrap(err)
		}
	}

	wrap = func(e error) error { return fmt.Errorf("invalid servers: %w", e) }
	if v := doc.Servers; v != nil {
		if err := v.Validate(ctx); err != nil {
			return wrap(err)
		}
	}

	wrap = func(e error) error { return fmt.Errorf("invalid tags: %w", e) }
	if v := doc.Tags; v != nil {
		if err := v.Validate(ctx); err != nil {
			return wrap(err)
		}
	}

	wrap = func(e error) error { return fmt.Errorf("invalid external docs: %w", e) }
	if v := doc.ExternalDocs; v != nil {
		if err := v.Validate(ctx); err != nil {
			return wrap(err)
		}
	}

	return nil
}
