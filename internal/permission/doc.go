// Package permission provides permission mapping between canonical policies and platform-specific formats.
//
// The Adapter interface previously declared here was removed (per the
// "interfaces where consumed" rule in golang-cli-architecture). The
// narrow contracts that consumers actually need are now declared in
// the consumer packages themselves:
//
//   - internal/parser.platformAdapter (needs ToCanonical)
//   - internal/renderer.templateAdapter (needs PermissionPolicyToPlatform, ConvertToolNameCase)
//
// Both internal/claude-code and internal/opencode satisfy these
// interfaces via Go structural typing — no compile-time tag or shim
// required.
package permission
