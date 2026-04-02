// Package library provides library management for canonical resources.
package library

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gitlab.com/amoconst/germinator/internal/domain"
	yaml "gopkg.in/yaml.v3"
)

// AddOptions contains options for adding a resource to the library.
type AddOptions struct {
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
func AddResource(opts AddOptions) error {
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
	lib, err := LoadLibrary(opts.LibraryPath)
	if err != nil {
		return fmt.Errorf("loading library: %w", err)
	}

	// Check for existing resource
	resourceKey := FormatRef(docType, name)
	existingTypeMap, typeExists := lib.Resources[docType]
	_, nameExists := existingTypeMap[name]
	if typeExists && nameExists && !opts.Force {
		return domain.NewFileError(opts.LibraryPath, "add", fmt.Sprintf("resource %s already exists (use --force to overwrite)", resourceKey), nil)
	}

	// Dry-run mode - report what would happen
	if opts.DryRun {
		fmt.Printf("Would add resource: %s\n", resourceKey)
		fmt.Printf("  Type: %s\n", docType)
		fmt.Printf("  Name: %s\n", name)
		fmt.Printf("  Description: %s\n", description)
		fmt.Printf("  Source: %s\n", opts.Source)
		return nil
	}

	// Determine target path
	targetDir := filepath.Join(opts.LibraryPath, docType+"s")
	targetFile := filepath.Join(targetDir, name+".md")

	// Ensure directory exists
	if err := os.MkdirAll(targetDir, 0o750); err != nil {
		return domain.NewFileError(targetDir, "create", "failed to create resource directory", err)
	}

	// Read source content
	content, err := os.ReadFile(opts.Source) //nolint:gosec // G304: User provides source path, must read user documents
	if err != nil {
		return domain.NewFileError(opts.Source, "read", "failed to read source file", err)
	}

	// Write to target path
	if err := os.WriteFile(targetFile, content, 0o644); err != nil { //nolint:gosec // G302: Creating new file with standard permissions
		return domain.NewFileError(targetFile, "write", "failed to write resource file", err)
	}

	// Update library.yaml
	if err := addResourceToLibrary(opts.LibraryPath, docType, name, targetFile, description); err != nil {
		return fmt.Errorf("updating library.yaml: %w", err)
	}

	// Validate the updated library
	if _, err := LoadLibrary(opts.LibraryPath); err != nil {
		return fmt.Errorf("validating updated library: %w", err)
	}

	return nil
}

// validateSourceExists checks if the source file exists and is readable.
func validateSourceExists(source string) error {
	info, err := os.Stat(source)
	if err != nil {
		if os.IsNotExist(err) {
			return domain.NewFileError(source, "access", "source file not found", nil)
		}
		return domain.NewFileError(source, "access", "failed to access source file", err)
	}
	if info.IsDir() {
		return domain.NewFileError(source, "access", "source path is a directory, expected a file", nil)
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
			return "", domain.NewConfigError("type", flag, "invalid resource type").WithSuggestions(validTypes)
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

	return "", domain.NewConfigError("type", "", "could not detect resource type (use --type flag or ensure file has type in frontmatter)")
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

	return "", domain.NewConfigError("name", "", "could not detect resource name (use --name flag or ensure file has name in frontmatter)")
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
		return domain.NewFileError(yamlPath, "read", "failed to read library.yaml", err)
	}

	var lib libraryYAML
	if err := yaml.Unmarshal(content, &lib); err != nil {
		return domain.NewParseError(yamlPath, "failed to parse library.yaml", err)
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
		return domain.NewParseError(yamlPath, "failed to marshal library.yaml", err)
	}

	// Write atomically: write to temp file first, then rename
	tmpPath := yamlPath + ".tmp"
	if err := os.WriteFile(tmpPath, output, 0o644); err != nil { //nolint:gosec // G302: Creating new file with standard permissions
		return domain.NewFileError(tmpPath, "write", "failed to write library.yaml", err)
	}
	if err := os.Rename(tmpPath, yamlPath); err != nil {
		return domain.NewFileError(yamlPath, "rename", "failed to update library.yaml", err)
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

// OrphanInfo represents an orphaned resource found during discovery.
type OrphanInfo struct {
	Path  string `json:"path"`
	Type  string `json:"type"`
	Name  string `json:"name"`
	Issue string `json:"issue,omitempty"` // "name_conflict" or empty
}

// ConflictInfo represents a conflict during orphan discovery.
type ConflictInfo struct {
	Orphan OrphanInfo `json:"orphan"`
	Issue  string     `json:"issue"`
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
	Orphans   []OrphanInfo    `json:"orphans"`
	Added     []AddSuccess    `json:"added"`
	Conflicts []ConflictInfo  `json:"conflicts"`
	Summary   DiscoverSummary `json:"summary"`
}

// DiscoverOrphans scans library directories for orphaned resource files.
func DiscoverOrphans(opts DiscoverOptions) (*DiscoverResult, error) {
	// Load the library to get registered resources
	lib, err := LoadLibrary(opts.LibraryPath)
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
		dirPath := filepath.Join(opts.LibraryPath, dir)
		if err := scanDirectory(dirPath, resType, lib, opts, result); err != nil {
			return nil, err
		}
	}

	// Update summary with totals
	result.Summary.TotalOrphans = len(result.Orphans)

	// Batch mode: process all orphans continuously, skipping errors
	if opts.Batch && opts.Force && !opts.DryRun {
		for _, orphan := range result.Orphans {
			if err := registerOrphan(opts.LibraryPath, orphan); err != nil {
				result.Summary.TotalFailed++
				continue
			}
			result.Added = append(result.Added, AddSuccess{
				Type: orphan.Type,
				Name: orphan.Name,
				Path: orphan.Path,
			})
			result.Summary.TotalAdded++
		}
	} else if opts.Force && len(result.Conflicts) == 0 && !opts.DryRun {
		// Legacy mode: require no conflicts before registering
		for _, orphan := range result.Orphans {
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

	return result, nil
}

// scanDirectory scans a single resource directory for orphans recursively.
func scanDirectory(dirPath, resType string, lib *Library, _ DiscoverOptions, result *DiscoverResult) error {
	err := filepath.WalkDir(dirPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			// Skip entries we can't access - continue walking
			return nil //nolint:wrapcheck // Intentional: continue walking on access errors
		}

		// Skip directories - we only process files
		if d.IsDir() {
			return nil
		}

		// Only process .md files
		if !strings.HasSuffix(strings.ToLower(path), ".md") {
			return nil
		}

		// Increment scanned count for .md files
		result.Summary.TotalScanned++

		orphan := detectOrphan(path, resType)

		// Check if already registered
		if isRegistered(lib, resType, orphan.Name) {
			return nil
		}

		// Check for name conflict with other type
		if hasNameConflict(lib, orphan.Name) {
			result.Conflicts = append(result.Conflicts, ConflictInfo{
				Orphan: orphan,
				Issue:  "name_conflict",
			})
			return nil
		}

		result.Orphans = append(result.Orphans, orphan)
		return nil
	})
	if err != nil {
		return fmt.Errorf("walking directory %s: %w", dirPath, err)
	}
	return nil
}

// detectOrphan detects orphan metadata from a file.
func detectOrphan(filePath, resType string) OrphanInfo {
	orphan := OrphanInfo{
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

// hasNameConflict checks if the orphan name conflicts with resources of other types.
func hasNameConflict(lib *Library, name string) bool {
	for _, resources := range lib.Resources {
		if _, exists := resources[name]; exists {
			return true
		}
	}
	return false
}

// registerOrphan adds an orphan to the library.
func registerOrphan(libraryPath string, orphan OrphanInfo) error {
	// Extract description from frontmatter when registering
	description := extractFrontmatterField(orphan.Path, "description")

	// Add to library.yaml using existing function
	err := addResourceToLibrary(libraryPath, orphan.Type, orphan.Name, orphan.Path, description)
	if err != nil {
		return err
	}

	return nil
}
