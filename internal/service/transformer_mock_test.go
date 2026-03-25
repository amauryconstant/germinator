package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/amoconst/germinator/internal/application"
	"gitlab.com/amoconst/germinator/internal/domain"
	"gitlab.com/amoconst/germinator/internal/infrastructure/parsing"
	"gitlab.com/amoconst/germinator/test/mocks"
)

func TestTransformerWithMocks(t *testing.T) {
	t.Run("success path uses injected parser and serializer", func(t *testing.T) {
		mockParser := new(mocks.MockParser)
		mockSerializer := new(mocks.MockSerializer)

		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "input.md")
		outputFile := filepath.Join(tmpDir, "output.md")

		// Set up mock expectations
		mockParser.On("LoadDocument", inputFile, "opencode").
			Return(&parsing.CanonicalAgent{Agent: domain.Agent{Name: "test-agent"}}, nil)
		mockSerializer.On("RenderDocument", mock.Anything, "opencode").
			Return("---\nname: test-agent\n---\nTransformed content", nil)

		tr := NewTransformer(mockParser, mockSerializer)

		result, err := tr.Transform(context.Background(), &application.TransformRequest{
			InputPath:  inputFile,
			OutputPath: outputFile,
			Platform:   "opencode",
		})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, outputFile, result.OutputPath)

		// Verify mocks were called
		mockParser.AssertExpectations(t)
		mockSerializer.AssertExpectations(t)

		// Verify output file was written
		content, err := os.ReadFile(outputFile)
		assert.NoError(t, err)
		assert.Contains(t, string(content), "Transformed content")
	})

	t.Run("parser error is propagated", func(t *testing.T) {
		mockParser := new(mocks.MockParser)
		mockSerializer := new(mocks.MockSerializer)

		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "input.md")
		outputFile := filepath.Join(tmpDir, "output.md")

		expectedErr := errors.New("parse error")
		mockParser.On("LoadDocument", inputFile, "opencode").
			Return(nil, expectedErr)

		tr := NewTransformer(mockParser, mockSerializer)

		result, err := tr.Transform(context.Background(), &application.TransformRequest{
			InputPath:  inputFile,
			OutputPath: outputFile,
			Platform:   "opencode",
		})

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "loading document")

		mockParser.AssertExpectations(t)
		mockSerializer.AssertNotCalled(t, "RenderDocument")

		// Verify output file was not created
		_, err = os.Stat(outputFile)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("serializer error is propagated", func(t *testing.T) {
		mockParser := new(mocks.MockParser)
		mockSerializer := new(mocks.MockSerializer)

		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "input.md")
		outputFile := filepath.Join(tmpDir, "output.md")

		mockParser.On("LoadDocument", inputFile, "claude-code").
			Return(&parsing.CanonicalAgent{Agent: domain.Agent{Name: "test-agent"}}, nil)
		mockSerializer.On("RenderDocument", mock.Anything, "claude-code").
			Return("", errors.New("render error"))

		tr := NewTransformer(mockParser, mockSerializer)

		result, err := tr.Transform(context.Background(), &application.TransformRequest{
			InputPath:  inputFile,
			OutputPath: outputFile,
			Platform:   "claude-code",
		})

		assert.Error(t, err)
		assert.Nil(t, result)

		mockParser.AssertExpectations(t)
		mockSerializer.AssertExpectations(t)
	})

	t.Run("write error returns FileError", func(t *testing.T) {
		mockParser := new(mocks.MockParser)
		mockSerializer := new(mocks.MockSerializer)

		inputFile := filepath.Join(t.TempDir(), "input.md")
		outputFile := "/nonexistent/directory/output.md"

		// Create input file
		if err := os.WriteFile(inputFile, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create input file: %v", err)
		}

		mockParser.On("LoadDocument", inputFile, "opencode").
			Return(&parsing.CanonicalAgent{Agent: domain.Agent{Name: "test"}}, nil)
		mockSerializer.On("RenderDocument", mock.Anything, "opencode").
			Return("rendered content", nil)

		tr := NewTransformer(mockParser, mockSerializer)

		result, err := tr.Transform(context.Background(), &application.TransformRequest{
			InputPath:  inputFile,
			OutputPath: outputFile,
			Platform:   "opencode",
		})

		assert.Error(t, err)
		assert.Nil(t, result)

		var fileErr *domain.FileError
		assert.True(t, errors.As(err, &fileErr))
		assert.Equal(t, "write", fileErr.Operation())
	})
}
