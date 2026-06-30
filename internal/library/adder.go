// Package library provides library management for canonical resources.
package library

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gitlab.com/amoconst/germinator/internal/core"
	yaml "gopkg.in/yaml.v3"
)

// ErrNameConflict is returned by [checkNameConflict] when an orphan name
// collides with an existing resource of a different type. It is the typed
// sentinel for the orphan-discovery name-conflict path; callers should use
// [errors.Is] to detect it in an error chain (for example, when wrapping
// it as the Cause of a [*core.OperationError]).
var ErrNameConflict = errors.New("name conflict with existing resource")

// AddRequest contains options for adding a resource to the library.
type AddRequest struct {
	// Source is the path to the source file to add (must be canonical format).
	Source string
	// Name is the optional resource name (overrides auto-detection).
	Name string
	// Description is the optional resource description (overrides auto-detection).
	Description string
	// Type is the optional resource type (overrides auto-detection).
	Type string
	// LibraryPath is the path to the library directory.
	LibraryPath string
	// Force overwrites an existing resource with the same name.
	Force bool
	// DryRun previews changes without modifying the library.
	DryRun bool
}

// AddResource adds a resource from a source file to the library.
// The source must already be in canonical format (canonicalization should be done by caller).
// It handles type detection, name detection, description detection, and library.yaml updates.
// The provided ctx is checked before each file I/O operation; on
// cancellation, the function returns wrapped ctx.Err() so the caller can
// distinguish a cancelled write from a regular library error.
func AddResource(ctx context.Context, opts AddRequest) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("adding resource: %w", err)
	}

	// Validate source file exists
	if err := validateSourceExists(opts.Source); err != nil {
		return err
	}

	// Detect resource type
	docType, err := detectType(opts.Source, opts.Type)
	if err != nil {
		return err
	}

	// Detect resource name
	name, err := detectName(opts.Source, opts.Name)
	if err != nil {
		return err
	}

	// Detect resource description
	description := detectDescription(opts.Source, opts.Description)

	// Load the library
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("adding resource: %w", err)
	}
	lib, err := LoadLibrary(ctx, opts.LibraryPath)
	if err != nil {
		return fmt.Errorf("loading library: %w", err)
	}

	// Check for existing resource
	resourceKey := FormatRef(docType, name)
	existingTypeMap, typeExists := lib.Resources[docType]
	_, nameExists := existingTypeMap[name]
	if typeExists && nameExists && !opts.Force {
		return core.NewFileError(opts.LibraryPath, "add", fmt.Sprintf("resource %s already exists (use --force to overwrite)", resourceKey), nil)
	}

	// Dry-run mode - report what would happen
	if opts.DryRun {
		_, _ = fmt.Fprintln(os.Stdout, "Would add resource:", resourceKey)
		_, _ = fmt.Fprintln(os.Stdout, "  Type:", docType)
		_, _ = fmt.Fprintln(os.Stdout, "  Name:", name)
		_, _ = fmt.Fprintln(os.Stdout, "  Description:", description)
		_, _ = fmt.Fprintln(os.Stdout, "  Source:", opts.Source)
		return nil
	}

	// Determine target path
	targetDir := filepath.Join(opts.LibraryPath, docType+"s")
	targetFile := filepath.Join(targetDir, name+".md")

	// Ensure directory exists
	if err := os.MkdirAll(targetDir, 0o750); err != nil {
		return core.NewFileError(targetDir, "create", "failed to create resource directory", err)
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("adding resource: %w", err)
	}

	// Read source content
	content, err := os.ReadFile(opts.Source) //nolint:gosec,nolintlint // G304: User provides source path, must read user documents
	if err != nil {
		return core.NewFileError(opts.Source, "read", "failed to read source file", err)
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("adding resource: %w", err)
	}

	// Write to target path
	if err := os.WriteFile(targetFile, content, 0o644); err != nil { //nolint:gosec // G302: Creating new file with standard permissions
		return core.NewFileError(targetFile, "write", "failed to write resource file", err)
	}

	// Update library.yaml
	if err := addResourceToLibrary(opts.LibraryPath, docType, name, targetFile, description); err != nil {
		return fmt.Errorf("updating library.yaml: %w", err)
	}

	// Validate the updated library
	if _, err := LoadLibrary(ctx, opts.LibraryPath); err != nil {
		return fmt.Errorf("validating updated library: %w", err)
	}

	return nil
}

// validateSourceExists checks if the source file exists and is readable.
func validateSourceExists(source string) error {
	info, err := os.Stat(source)
	if err != nil {
		if os.IsNotExist(err) {
			return core.NewFileError(source, "access", "source file not found", nil)
		}
		return core.NewFileError(source, "access", "failed to access source file", err)
	}
	if info.IsDir() {
		return core.NewFileError(source, "access", "source path is a directory, expected a file", nil)
	}
	return nil
}

// detectType detects the resource type from flag, frontmatter, or filename.
// Priority: flag > frontmatter > filename
func detectType(source, flag string) (string, error) {
	// 1. Flag takes precedence
	if flag != "" {
		rt := ResourceType(flag)
		if !rt.IsValid() {
			validTypes := make([]string, len(ValidResourceTypes))
			for i, t := range ValidResourceTypes {
				validTypes[i] = string(t)
			}
			return "", core.NewConfigError("type", flag, "invalid resource type").WithSuggestions(validTypes)
		}
		return flag, nil
	}

	// 2. Try to detect from frontmatter
	if docType := extractFrontmatterField(source, "type"); docType != "" {
		rt := ResourceType(docType)
		if rt.IsValid() {
			return docType, nil
		}
	}

	// 3. Fallback to filename pattern
	if docType := DetectTypeFromFilename(source); docType != "" {
		return docType, nil
	}

	return "", core.NewConfigError("type", "", "could not detect resource type (use --type flag or ensure file has type in frontmatter)")
}

// detectName detects the resource name from flag or frontmatter.
// Priority: flag > frontmatter > filename
func detectName(source, flag string) (string, error) {
	// 1. Flag takes precedence
	if flag != "" {
		return flag, nil
	}

	// 2. Try to detect from frontmatter
	if name := extractFrontmatterField(source, "name"); name != "" {
		return name, nil
	}

	// 3. Try filename (e.g., "skill-commit.md" -> "commit")
	if name := extractNameFromFilename(source); name != "" {
		return name, nil
	}

	return "", core.NewConfigError("name", "", "could not detect resource name (use --name flag or ensure file has name in frontmatter)")
}

// detectDescription detects the resource description from flag or frontmatter.
// Priority: flag > frontmatter
func detectDescription(source, flag string) string {
	// 1. Flag takes precedence
	if flag != "" {
		return flag
	}

	// 2. Try to detect from frontmatter
	if desc := extractFrontmatterField(source, "description"); desc != "" {
		return desc
	}

	// Description is optional - return empty string
	return ""
}

// DetectPlatform detects the platform from frontmatter or filename.
// Returns empty string if cannot detect (assumes canonical).
func DetectPlatform(source string) string {
	// Check frontmatter for platform field
	if platform := extractFrontmatterField(source, "platform"); platform != "" {
		if IsValidPlatform(platform) {
			return platform
		}
	}

	// Check filename for platform indicators
	lower := strings.ToLower(source)
	if strings.Contains(lower, "opencode") {
		return "opencode"
	}
	if strings.Contains(lower, "claude-code") || strings.Contains(lower, "claudecode") {
		return "claude-code"
	}

	return ""
}

// IsCanonicalFormat checks if the source file is already in canonical format.
func IsCanonicalFormat(source, docType string) bool {
	yamlContent, err := extractFrontmatter(source)
	if err != nil || yamlContent == "" {
		return false
	}

	var frontmatter map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlContent), &frontmatter); err != nil {
		return false
	}

	// Canonical format has "name", "description" fields
	// Platform format has additional fields like "allowed-tools", "mode", "permissionMode"
	if _, hasName := frontmatter["name"]; !hasName {
		return false
	}
	if _, hasDesc := frontmatter["description"]; !hasDesc {
		return false
	}

	// Check for platform-specific fields
	switch docType {
	case "command", "skill":
		if _, ok := frontmatter["allowed-tools"]; ok {
			return false // OpenCode format
		}
	case "agent":
		if _, ok := frontmatter["permissionMode"]; ok {
			return false // Claude Code format
		}
		if _, ok := frontmatter["mode"]; ok {
			return false // OpenCode format
		}
	}

	return true
}

// addResourceToLibrary adds a resource entry to library.yaml.
func addResourceToLibrary(libraryPath, docType, name, filePath, description string) error {
	yamlPath := filepath.Join(libraryPath, "library.yaml")

	// Read current library.yaml
	content, err := os.ReadFile(yamlPath) //nolint:gosec // G304: User provides library path, must read fixed library.yaml
	if err != nil {
		return core.NewFileError(yamlPath, "read", "failed to read library.yaml", err)
	}

	var lib libraryYAML
	if err := yaml.Unmarshal(content, &lib); err != nil {
		return core.NewParseError(yamlPath, "failed to parse library.yaml", err)
	}

	// Initialize maps if nil
	if lib.Resources == nil {
		lib.Resources = make(map[string]map[string]Resource)
	}
	if lib.Resources[docType] == nil {
		lib.Resources[docType] = make(map[string]Resource)
	}

	// Compute relative path from library root
	relPath, err := filepath.Rel(libraryPath, filePath)
	if err != nil {
		return fmt.Errorf("computing relative path: %w", err)
	}

	// Add/update resource entry
	lib.Resources[docType][name] = Resource{
		Path:        relPath,
		Description: description,
	}

	// Marshal back to YAML
	output, err := yaml.Marshal(lib)
	if err != nil {
		return core.NewParseError(yamlPath, "failed to marshal library.yaml", err)
	}

	// Write atomically: write to temp file first, then rename
	tmpPath := yamlPath + ".tmp"
	if err := os.WriteFile(tmpPath, output, 0o644); err != nil { //nolint:gosec // G302: Creating new file with standard permissions
		return core.NewFileError(tmpPath, "write", "failed to write library.yaml", err)
	}
	if err := os.Rename(tmpPath, yamlPath); err != nil {
		return core.NewFileError(yamlPath, "rename", "failed to update library.yaml", err)
	}

	return nil
}

// extractFrontmatter extracts YAML frontmatter from a markdown file.
// Returns the YAML content or empty string if no frontmatter.
func extractFrontmatter(source string) (string, error) {
	content, err := os.ReadFile(source) //nolint:gosec // G304: User provides source path, must read user documents
	if err != nil {
		return "", fmt.Errorf("reading file: %w", err)
	}

	fileContent := string(content)
	lines := strings.Split(fileContent, "\n")
	if len(lines) < 3 || lines[0] != "---" {
		return "", nil
	}

	var yamlLines []string
	foundEnd := false

	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			foundEnd = true
			break
		}
		yamlLines = append(yamlLines, lines[i])
	}

	if !foundEnd {
		return "", nil
	}

	return strings.Join(yamlLines, "\n"), nil
}

// extractFrontmatterField extracts a specific string field from YAML frontmatter.
func extractFrontmatterField(source, field string) string {
	yamlContent, err := extractFrontmatter(source)
	if err != nil || yamlContent == "" {
		return ""
	}

	var frontmatter map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlContent), &frontmatter); err != nil {
		return ""
	}

	if value, ok := frontmatter[field].(string); ok {
		return value
	}

	return ""
}

// DetectTypeFromFilename detects document type from filename patterns.
func DetectTypeFromFilename(filepath string) string {
	base := filepath

	// Check agent patterns
	if matched, _ := regexp.MatchString(`agent-.*\..*$`, base); matched {
		return "agent"
	}
	if matched, _ := regexp.MatchString(`.*-agent\..*$`, base); matched {
		return "agent"
	}

	// Check command patterns
	if matched, _ := regexp.MatchString(`command-.*\..*$`, base); matched {
		return "command"
	}
	if matched, _ := regexp.MatchString(`.*-command\..*$`, base); matched {
		return "command"
	}

	// Check memory patterns
	if matched, _ := regexp.MatchString(`memory-.*\..*$`, base); matched {
		return "memory"
	}
	if matched, _ := regexp.MatchString(`.*-memory\..*$`, base); matched {
		return "memory"
	}

	// Check skill patterns
	if matched, _ := regexp.MatchString(`skill-.*\..*$`, base); matched {
		return "skill"
	}
	if matched, _ := regexp.MatchString(`.*-skill\..*$`, base); matched {
		return "skill"
	}

	return ""
}

// extractNameFromFilename extracts a resource name from filename.
// E.g., "agent-reviewer.md" -> "reviewer", "skill-commit.md" -> "commit"
func extractNameFromFilename(source string) string {
	// Get base filename without extension
	base := filepath.Base(source)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	// Remove prefix patterns like "agent-", "skill-", "command-", "memory-"
	patterns := []string{
		`^agent[-_]`,
		`^skill[-_]`,
		`^command[-_]`,
		`^memory[-_]`,
		`[-_]agent$`,
		`[-_]skill$`,
		`[-_]command$`,
		`[-_]memory$`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		name = re.ReplaceAllString(name, "")
	}

	return name
}

// DiscoverOptions contains options for discovering orphaned resources.
type DiscoverOptions struct {
	LibraryPath string
	DryRun      bool
	Force       bool
	Batch       bool
}

// Orphan represents an orphaned resource found during discovery.
type Orphan struct {
	Path  string `json:"path"`
	Type  string `json:"type"`
	Name  string `json:"name"`
	Issue string `json:"issue,omitempty"` // "name_conflict" or empty
}

// ConflictInfo represents a conflict during orphan discovery.
type ConflictInfo struct {
	Orphan Orphan `json:"orphan"`
	Issue  string `json:"issue"`
	Cause  error  `json:"-"`
}

// AddSuccess represents a successfully added orphan resource.
type AddSuccess struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Path string `json:"path"`
}

// DiscoverSummary contains statistics from an orphan discovery operation.
type DiscoverSummary struct {
	TotalScanned int `json:"totalScanned"`
	TotalOrphans int `json:"totalOrphans"`
	TotalAdded   int `json:"totalAdded"`
	TotalSkipped int `json:"totalSkipped"`
	TotalFailed  int `json:"totalFailed"`
}

// DiscoverResult contains the result of an orphan discovery operation.
type DiscoverResult struct {
	Orphans   []Orphan        `json:"orphans"`
	Added     []AddSuccess    `json:"added"`
	Conflicts []ConflictInfo  `json:"conflicts"`
	Summary   DiscoverSummary `json:"summary"`
}

// BatchAddResult contains the result of a batch add operation.
type BatchAddResult struct {
	Added   []BatchAddSuccess  `json:"added"`
	Skipped []BatchSkipInfo    `json:"skipped"`
	Failed  []BatchFailureInfo `json:"failed"`
	Summary BatchSummary       `json:"summary"`
}

// BatchAddSuccess represents a successfully added resource in batch mode.
type BatchAddSuccess struct {
	Ref  string `json:"ref"`  // Resource reference (e.g., "skill/commit")
	Path string `json:"path"` // Path in library
}

// BatchSkipInfo represents a resource that was skipped in batch mode.
type BatchSkipInfo struct {
	Source string `json:"source"` // Original source path
	Issue  string `json:"issue"`  // "already_exists" or "conflict"
}

// BatchFailureInfo represents a failed resource add in batch mode.
type BatchFailureInfo struct {
	Source string `json:"source"` // Original source path
	Error  string `json:"error"`  // Error message
}

// BatchSummary contains statistics from a batch add operation.
type BatchSummary struct {
	Total   int `json:"total"`
	Added   int `json:"added"`
	Skipped int `json:"skipped"`
	Failed  int `json:"failed"`
}

// BatchAddOptions contains options for batch adding resources.
type BatchAddOptions struct {
	Sources     []string // Source files/directories to add
	LibraryPath string   // Path to the library
	DryRun      bool     // Preview without modifying
	Force       bool     // Overwrite existing resources
	Name        string   // Optional resource name override
	Description string   // Optional resource description override
	Type        string   // Optional resource type override
	Platform    string   // Optional platform override
	Orphans     []Orphan // Orphan info for discovered resources (provides type/name)
}

// BatchAddResources adds multiple resources to the library in batch mode.
// It processes all sources sequentially, collecting results by category
// (added/skipped/failed). The provided ctx is checked between files in the
// inner loop; on cancellation the partial BatchAddResult is returned
// alongside wrapped ctx.Err() so callers can inspect successes/failures
// observed up to the cancel point.
func BatchAddResources(ctx context.Context, opts BatchAddOptions) (*BatchAddResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("batch add: %w", err)
	}

	result := &BatchAddResult{}

	// Build a map of source path to orphan info for quick lookup
	orphanMap := make(map[string]Orphan)
	for _, orphan := range opts.Orphans {
		orphanMap[orphan.Path] = orphan
	}

	// Collect all source files (expand directories)
	files, err := collectSourceFiles(opts.Sources)
	if err != nil {
		return nil, fmt.Errorf("collecting source files: %w", err)
	}

	// Initialize summary total
	result.Summary.Total = len(files)

	// Process each file
	for _, source := range files {
		if cerr := ctx.Err(); cerr != nil {
			return result, fmt.Errorf("batch add: %w", cerr)
		}
		var orphan Orphan
		if o, ok := orphanMap[source]; ok {
			orphan = o
		}
		_ = processBatchAddFile(ctx, source, opts, result, orphan) // Error is already recorded in result
	}

	if err := ctx.Err(); err != nil {
		return result, fmt.Errorf("batch add: %w", err)
	}

	return result, nil
}

// collectSourceFiles expands directories to .md files recursively.
func collectSourceFiles(sources []string) ([]string, error) {
	var files []string

	for _, source := range sources {
		info, err := os.Stat(source)
		if err != nil {
			if os.IsNotExist(err) {
				// Treat non-existent paths as single files (will fail later with proper error)
				files = append(files, source)
				continue
			}
			return nil, fmt.Errorf("accessing %s: %w", source, err)
		}

		if info.IsDir() {
			// Walk directory tree to find all .md files
			err := filepath.WalkDir(source, func(path string, d os.DirEntry, walkErr error) error {
				if walkErr != nil {
					return nil // Skip entries we can't access
				}
				if d.IsDir() {
					return nil
				}
				if strings.HasSuffix(strings.ToLower(path), ".md") {
					files = append(files, path)
				}
				return nil
			})
			if err != nil {
				return nil, fmt.Errorf("walking directory %s: %w", source, err)
			}
		} else {
			files = append(files, source)
		}
	}

	return files, nil
}

// processBatchAddFile processes a single file for batch add.
// The ctx is propagated into the library load and the recursive AddResource
// call; a mid-batch cancellation is reported as a per-file failure so the
// partial summary remains consistent.
func processBatchAddFile(ctx context.Context, source string, opts BatchAddOptions, result *BatchAddResult, orphan Orphan) error {
	// Pre-flight cancellation check so a cancelled context short-circuits
	// the detect → load → add pipeline per file.
	if err := ctx.Err(); err != nil {
		result.Failed = append(result.Failed, BatchFailureInfo{
			Source: source,
			Error:  err.Error(),
		})
		result.Summary.Failed++
		return fmt.Errorf("processBatchAddFile: %w", err)
	}

	// Detect resource type - use orphan type if available (from discover), otherwise detect
	docType := orphan.Type
	if docType == "" {
		docType = opts.Type
	}
	if docType == "" {
		var err error
		docType, err = detectType(source, opts.Type)
		if err != nil {
			result.Failed = append(result.Failed, BatchFailureInfo{
				Source: source,
				Error:  err.Error(),
			})
			result.Summary.Failed++
			return err
		}
	}

	// Detect resource name - use orphan name if available
	name := orphan.Name
	if name == "" {
		name = opts.Name
	}
	if name == "" {
		var err error
		name, err = detectName(source, opts.Name)
		if err != nil {
			result.Failed = append(result.Failed, BatchFailureInfo{
				Source: source,
				Error:  err.Error(),
			})
			result.Summary.Failed++
			return err
		}
	}

	// Detect resource description (from frontmatter or opts)
	description := detectDescription(source, opts.Description)

	// Format reference
	resourceKey := FormatRef(docType, name)

	// Load the library to check for existing resources
	lib, err := LoadLibrary(ctx, opts.LibraryPath)
	if err != nil {
		result.Failed = append(result.Failed, BatchFailureInfo{
			Source: source,
			Error:  fmt.Sprintf("loading library: %v", err),
		})
		result.Summary.Failed++
		return fmt.Errorf("loading library: %w", err)
	}

	// Check for existing resource (skip check if Force is set)
	existingTypeMap, typeExists := lib.Resources[docType]
	_, nameExists := existingTypeMap[name]
	if typeExists && nameExists && !opts.Force {
		result.Skipped = append(result.Skipped, BatchSkipInfo{
			Source: source,
			Issue:  "already_exists",
		})
		result.Summary.Skipped++
		return nil
	}

	// Check for name conflict with other types (skip check if Force is set)
	hasConflict := false
	if !opts.Force {
		for rType, resources := range lib.Resources {
			if rType == docType {
				continue
			}
			if _, exists := resources[name]; exists {
				hasConflict = true
				break
			}
		}
	}
	if hasConflict {
		result.Skipped = append(result.Skipped, BatchSkipInfo{
			Source: source,
			Issue:  "conflict",
		})
		result.Summary.Skipped++
		return nil
	}

	// Dry-run mode - record what would be added
	if opts.DryRun {
		result.Added = append(result.Added, BatchAddSuccess{
			Ref:  resourceKey,
			Path: filepath.Join(opts.LibraryPath, docType+"s", name+".md"),
		})
		result.Summary.Added++
		return nil
	}

	// Add the resource using existing AddResource function
	addErr := AddResource(ctx, AddRequest{
		Source:      source,
		Name:        name,
		Description: description,
		Type:        docType,
		LibraryPath: opts.LibraryPath,
		Force:       opts.Force,
		DryRun:      false,
	})

	if addErr != nil {
		result.Failed = append(result.Failed, BatchFailureInfo{
			Source: source,
			Error:  addErr.Error(),
		})
		result.Summary.Failed++
		return addErr
	}

	result.Added = append(result.Added, BatchAddSuccess{
		Ref:  resourceKey,
		Path: filepath.Join(opts.LibraryPath, docType+"s", name+".md"),
	})
	result.Summary.Added++
	return nil
}

// DiscoverOrphans scans library directories for orphaned resource files.
// The provided ctx is checked between scanned directories and inside the
// per-directory walk; on cancellation, the partial DiscoverResult is
// returned alongside wrapped ctx.Err() so callers can inspect what was
// found before the cancel arrived.
func DiscoverOrphans(ctx context.Context, opts DiscoverOptions) (*DiscoverResult, error) {
	// Load the library to get registered resources
	lib, err := LoadLibrary(ctx, opts.LibraryPath)
	if err != nil {
		return nil, fmt.Errorf("loading library: %w", err)
	}

	result := &DiscoverResult{}

	// Scan each resource directory
	directories := map[string]string{
		"skills":   "skill",
		"agents":   "agent",
		"commands": "command",
		"memory":   "memory",
	}

	for dir, resType := range directories {
		if err := ctx.Err(); err != nil {
			return result, fmt.Errorf("discovering orphans: %w", err)
		}
		dirPath := filepath.Join(opts.LibraryPath, dir)
		if err := scanDirectory(ctx, dirPath, resType, lib, opts, result); err != nil {
			return result, err
		}
	}

	// Update summary with totals
	result.Summary.TotalOrphans = len(result.Orphans)

	// Legacy non-batch force mode: require no conflicts before registering
	// Note: When Batch is true, we defer to BatchAddResources for full processing
	if !opts.Batch && opts.Force && len(result.Conflicts) == 0 && !opts.DryRun {
		for _, orphan := range result.Orphans {
			if err := ctx.Err(); err != nil {
				return result, fmt.Errorf("discovering orphans: %w", err)
			}
			if err := registerOrphan(opts.LibraryPath, orphan); err != nil {
				return nil, err
			}
			result.Added = append(result.Added, AddSuccess{
				Type: orphan.Type,
				Name: orphan.Name,
				Path: orphan.Path,
			})
			result.Summary.TotalAdded++
		}
	}

	if err := ctx.Err(); err != nil {
		return result, fmt.Errorf("discovering orphans: %w", err)
	}

	return result, nil
}

// scanDirectory scans a single resource directory for orphans recursively.
// The walker inspects ctx.Err() per file entry so that a mid-walk
// cancellation surfaces a wrapped context.Canceled / DeadlineExceeded
// instead of running to completion.
func scanDirectory(ctx context.Context, dirPath, resType string, lib *Library, _ DiscoverOptions, result *DiscoverResult) error {
	err := filepath.WalkDir(dirPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			// Skip entries we can't access - continue walking
			return nil
		}

		// Skip directories - we only process files
		if d.IsDir() {
			return nil
		}

		// Only process .md files
		if !strings.HasSuffix(strings.ToLower(path), ".md") {
			return nil
		}

		// Per-file cancellation check inside the walk.
		if cerr := ctx.Err(); cerr != nil {
			return fmt.Errorf("scan: %w", cerr)
		}

		// Increment scanned count for .md files
		result.Summary.TotalScanned++

		orphan := detectOrphan(path, resType)

		// Check if already registered
		if isRegistered(lib, resType, orphan.Name) {
			return nil
		}

		// Check for name conflict with other type. checkNameConflict
		// returns ErrNameConflict wrapped with the offending <type>/<name>
		// ref so callers (notably runAdd Mode 2 in task 6.4) can wrap the
		// collision as the Cause of a *core.OperationError. We surface it
		// via the per-file ConflictInfo here so the existing JSON / human
		// output continues to show the conflict alongside the orphan.
		if conflictErr := checkNameConflict(lib, &orphan); conflictErr != nil {
			result.Conflicts = append(result.Conflicts, ConflictInfo{
				Orphan: orphan,
				Issue:  conflictErr.Error(),
				Cause:  conflictErr,
			})
			return nil
		}

		result.Orphans = append(result.Orphans, orphan)
		return nil
	})
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("scanning directory %s: %w", dirPath, err)
		}
		return fmt.Errorf("walking directory %s: %w", dirPath, err)
	}
	return nil
}

// detectOrphan detects orphan metadata from a file.
func detectOrphan(filePath, resType string) Orphan {
	orphan := Orphan{
		Type: resType,
		Path: filePath,
	}

	// Try to get name from frontmatter first, then filename
	if name := extractFrontmatterField(filePath, "name"); name != "" {
		orphan.Name = name
	} else {
		orphan.Name = extractNameFromFilename(filePath)
	}

	return orphan
}

// isRegistered checks if a resource is already registered in the library.
func isRegistered(lib *Library, resType, name string) bool {
	typeMap, exists := lib.Resources[resType]
	if !exists {
		return false
	}
	_, exists = typeMap[name]
	return exists
}

// checkNameConflict reports whether the given orphan collides with an
// already-registered resource of a different type. On collision it
// returns ErrNameConflict wrapped with the offending "<type>/<name>"
// ref (e.g., "agent/commit: name conflict with existing resource"); on
// no collision it returns nil.
//
// Cross-type only by design: same-type duplicates are surfaced earlier
// by isRegistered (and gated on --force by AddResource), so they are not
// reported as a typed conflict here. This keeps "the same type already
// has this name" distinct from "a different type already has this name"
// for callers that branch on errors.Is(err, ErrNameConflict).
func checkNameConflict(lib *Library, orphan *Orphan) error {
	for resType, resources := range lib.Resources {
		if resType == orphan.Type {
			continue
		}
		if _, exists := resources[orphan.Name]; exists {
			return fmt.Errorf("%s/%s: %w", orphan.Type, orphan.Name, ErrNameConflict)
		}
	}
	return nil
}

// registerOrphan adds an orphan to the library.
func registerOrphan(libraryPath string, orphan Orphan) error {
	// Extract description from frontmatter when registering
	description := extractFrontmatterField(orphan.Path, "description")

	// Add to library.yaml using existing function
	err := addResourceToLibrary(libraryPath, orphan.Type, orphan.Name, orphan.Path, description)
	if err != nil {
		return err
	}

	return nil
}
