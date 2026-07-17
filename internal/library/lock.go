package library

import (
	"path/filepath"
	"time"

	"github.com/gofrs/flock"

	gerrors "gitlab.com/amoconst/germinator/internal/core"
)

// lockRetryInterval is the gap between TryLock attempts when another
// process holds library.lock.
const lockRetryInterval = 25 * time.Millisecond

// lockMaxWait is the total budget for acquiring library.lock before
// returning *core.FileError(op="lock") to the caller. CLIs are
// short-lived; a 250 ms budget covers reasonable contention without
// blocking the user's shell.
const lockMaxWait = 250 * time.Millisecond

// withFileLock acquires an exclusive flock on <libraryPath>/library.lock,
// runs fn, and releases the lock when fn returns (success or error).
// The lock file lives alongside library.yaml; it is not removed on
// release because holding the file is harmless and re-creating it on
// every call is wasteful.
//
// On contention with another writer, withFileLock retries TryLock
// every lockRetryInterval until lockMaxWait elapses, then returns
// *core.FileError(path, op="lock", message, cause) so the CLI can
// render "another writer is active" via output.FormatError.
//
// Lock scope rationale: the lock spans the full read-modify-write
// cycle for each mutating entry point (SaveLibrary, addResourceToLibrary,
// removeResourceFromLibrary, removePresetFromLibrary, RefreshLibrary's
// load→process→save cycle) so two concurrent `germinator library …`
// invocations cannot interleave reads and writes — the second process
// either waits up to 250 ms or surfaces a clear error. Plain
// temp+rename atomicity (atomicWriteFile) only protects against
// crashes, not against dual-process writers; references/10-state.md
// §"Concurrent Invocation" calls this out explicitly.
func withFileLock(libraryPath string, fn func() error) error {
	if libraryPath == "" {
		return gerrors.NewFileError("", "lock", "library path is empty", nil)
	}
	lockPath := filepath.Join(libraryPath, "library.lock")
	lock := flock.New(lockPath)
	defer func() {
		// Unlock is idempotent on an unacquired lock — safe to defer
		// before the TryLock loop returns the timeout error.
		_ = lock.Unlock()
	}()

	deadline := time.Now().Add(lockMaxWait)
	for {
		acquired, err := lock.TryLock()
		if err != nil {
			return gerrors.NewFileError(lockPath, "lock", "failed to acquire file lock", err)
		}
		if acquired {
			return fn()
		}
		if time.Now().After(deadline) {
			return gerrors.NewFileError(lockPath, "lock",
				"another writer is active (lock contention timeout)", nil)
		}
		time.Sleep(lockRetryInterval)
	}
}
