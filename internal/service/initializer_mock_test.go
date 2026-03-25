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
	"gitlab.com/amoconst/germinator/internal/infrastructure/library"
	"gitlab.com/amoconst/germinator/internal/infrastructure/parsing"
	"gitlab.com/amoconst/germinator/test/mocks"
)

func TestInitializerWithMocks(t *testing.T) {
	t.Run("success path uses injected parser and serializer", func(t *testing.T) {
		mockParser := new(mocks.MockParser)
		mockSerializer := new(mocks.MockSerializer)

		// Create a minimal library
		tmpDir := t.TempDir()
		libRoot := filepath.Join(tmpDir, "lib")
		if err := os.MkdirAll(filepath.Join(libRoot, "skills", "commit"), 0755); err != nil {
			t.Fatalf("Failed to create library directory: %v", err)
		}
		// Create a minimal skill file
		skillContent := `---
name: commit
description: Git commit skill
---
Commit content`
		if err := os.WriteFile(filepath.Join(libRoot, "skills", "commit", "skill-commit.md"), []byte(skillContent), 0644); err != nil {
			t.Fatalf("Failed to write skill file: %v", err)
		}

		lib := &library.Library{
			RootPath: libRoot,
			Resources: map[string]map[string]library.Resource{
				"skill": {
					"commit": {Path: filepath.Join("skills", "commit", "skill-commit.md")},
				},
			},
			Presets: map[string]library.Preset{},
		}

		outputDir := t.TempDir()

		mockParser.On("LoadDocument", mock.AnythingOfType("string"), "opencode").
			Return(&parsing.CanonicalSkill{Skill: domain.Skill{Name: "commit"}}, nil)
		mockSerializer.On("RenderDocument", mock.Anything, "opencode").
			Return("--- skill content ---\n", nil)

		init := NewInitializer(mockParser, mockSerializer)

		results, err := init.Initialize(context.Background(), &application.InitializeRequest{
			Library:   lib,
			Platform:  "opencode",
			OutputDir: outputDir,
			Refs:      []string{"skill/commit"},
			DryRun:    false,
			Force:     true,
		})

		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "skill/commit", results[0].Ref)
		assert.NoError(t, results[0].Error)

		mockParser.AssertExpectations(t)
		mockSerializer.AssertExpectations(t)
	})

	t.Run("parser error propagates error", func(t *testing.T) {
		mockParser := new(mocks.MockParser)
		mockSerializer := new(mocks.MockSerializer)

		tmpDir := t.TempDir()
		libRoot := filepath.Join(tmpDir, "lib")
		if err := os.MkdirAll(filepath.Join(libRoot, "skills", "commit"), 0755); err != nil {
			t.Fatalf("Failed to create library directory: %v", err)
		}
		skillContent := `---
name: commit
description: Git commit skill
---
Commit content`
		if err := os.WriteFile(filepath.Join(libRoot, "skills", "commit", "skill-commit.md"), []byte(skillContent), 0644); err != nil {
			t.Fatalf("Failed to write skill file: %v", err)
		}

		lib := &library.Library{
			RootPath: libRoot,
			Resources: map[string]map[string]library.Resource{
				"skill": {
					"commit": {Path: filepath.Join("skills", "commit", "skill-commit.md")},
				},
			},
			Presets: map[string]library.Preset{},
		}

		outputDir := t.TempDir()

		expectedErr := errors.New("parse error")
		mockParser.On("LoadDocument", mock.AnythingOfType("string"), "opencode").
			Return(nil, expectedErr)

		init := NewInitializer(mockParser, mockSerializer)

		results, err := init.Initialize(context.Background(), &application.InitializeRequest{
			Library:   lib,
			Platform:  "opencode",
			OutputDir: outputDir,
			Refs:      []string{"skill/commit"},
			DryRun:    false,
			Force:     true,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "loading document")
		assert.Len(t, results, 1)
		assert.Error(t, results[0].Error)

		mockParser.AssertExpectations(t)
		mockSerializer.AssertNotCalled(t, "RenderDocument")
	})

	t.Run("serializer error propagates error", func(t *testing.T) {
		mockParser := new(mocks.MockParser)
		mockSerializer := new(mocks.MockSerializer)

		tmpDir := t.TempDir()
		libRoot := filepath.Join(tmpDir, "lib")
		if err := os.MkdirAll(filepath.Join(libRoot, "skills", "commit"), 0755); err != nil {
			t.Fatalf("Failed to create library directory: %v", err)
		}
		skillContent := `---
name: commit
description: Git commit skill
---
Commit content`
		if err := os.WriteFile(filepath.Join(libRoot, "skills", "commit", "skill-commit.md"), []byte(skillContent), 0644); err != nil {
			t.Fatalf("Failed to write skill file: %v", err)
		}

		lib := &library.Library{
			RootPath: libRoot,
			Resources: map[string]map[string]library.Resource{
				"skill": {
					"commit": {Path: filepath.Join("skills", "commit", "skill-commit.md")},
				},
			},
			Presets: map[string]library.Preset{},
		}

		outputDir := t.TempDir()

		mockParser.On("LoadDocument", mock.AnythingOfType("string"), "opencode").
			Return(&parsing.CanonicalSkill{Skill: domain.Skill{Name: "commit"}}, nil)
		mockSerializer.On("RenderDocument", mock.Anything, "opencode").
			Return("", errors.New("render error"))

		init := NewInitializer(mockParser, mockSerializer)

		results, err := init.Initialize(context.Background(), &application.InitializeRequest{
			Library:   lib,
			Platform:  "opencode",
			OutputDir: outputDir,
			Refs:      []string{"skill/commit"},
			DryRun:    false,
			Force:     true,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rendering document")
		assert.Len(t, results, 1)
		assert.Error(t, results[0].Error)

		mockParser.AssertExpectations(t)
		mockSerializer.AssertExpectations(t)
	})

	t.Run("dry-run does not call parser or serializer", func(t *testing.T) {
		mockParser := new(mocks.MockParser)
		mockSerializer := new(mocks.MockSerializer)

		tmpDir := t.TempDir()
		libRoot := filepath.Join(tmpDir, "lib")
		if err := os.MkdirAll(filepath.Join(libRoot, "skills", "commit"), 0755); err != nil {
			t.Fatalf("Failed to create library directory: %v", err)
		}
		skillContent := `---
name: commit
description: Git commit skill
---
Commit content`
		if err := os.WriteFile(filepath.Join(libRoot, "skills", "commit", "skill-commit.md"), []byte(skillContent), 0644); err != nil {
			t.Fatalf("Failed to write skill file: %v", err)
		}

		lib := &library.Library{
			RootPath: libRoot,
			Resources: map[string]map[string]library.Resource{
				"skill": {
					"commit": {Path: filepath.Join("skills", "commit", "skill-commit.md")},
				},
			},
			Presets: map[string]library.Preset{},
		}

		outputDir := t.TempDir()

		init := NewInitializer(mockParser, mockSerializer)

		results, err := init.Initialize(context.Background(), &application.InitializeRequest{
			Library:   lib,
			Platform:  "opencode",
			OutputDir: outputDir,
			Refs:      []string{"skill/commit"},
			DryRun:    true,
			Force:     false,
		})

		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "skill/commit", results[0].Ref)

		mockParser.AssertNotCalled(t, "LoadDocument")
		mockSerializer.AssertNotCalled(t, "RenderDocument")
	})
}
