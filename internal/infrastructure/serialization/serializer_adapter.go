package serialization

import (
	"gitlab.com/amoconst/germinator/internal/application"
)

// serializationSerializer implements the application.Serializer interface.
type serializationSerializer struct{}

// NewSerializer creates a new Serializer instance that implements the application.Serializer interface.
func NewSerializer() application.Serializer {
	return &serializationSerializer{}
}

// RenderDocument renders a document to the target platform format.
func (s *serializationSerializer) RenderDocument(doc interface{}, platform string) (string, error) {
	return RenderDocument(doc, platform)
}

// Compile-time interface satisfaction check.
var _ application.Serializer = (*serializationSerializer)(nil)
