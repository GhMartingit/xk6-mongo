package file

import (
	"context"
	"errors"
	"testing"
	"time"
)

func Test_TryLock(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// this is the original lock
	firstLock := newDirLock(dir)

	// should lock dir without errors
	if err := firstLock.tryLock(); err != nil {
		t.Fatalf("unexpected %v", err)
	}

	//  locking again should return without errors
	if err := firstLock.tryLock(); err != nil {
		t.Fatalf("unexpected %v", err)
	}

	// another lock should return ErrLocked
	if err := newDirLock(dir).tryLock(); !errors.Is(err, errLocked) {
		t.Fatalf("unexpected %v", err)
	}

	// locking another directory return without errors
	anotherLock := newDirLock(t.TempDir())
	if err := anotherLock.tryLock(); err != nil {
		t.Fatalf("unexpected %v", err)
	}
	// must unlock or test can't clean up the tmp dir
	defer anotherLock.unlock() //nolint:errcheck

	// unlock should work
	if err := firstLock.unlock(); err != nil {
		t.Fatalf("unexpected %v", err)
	}

	// unlocking again should return without errors
	if err := firstLock.unlock(); err != nil {
		t.Fatalf("unexpected %v", err)
	}

	// trying another lock again should work now
	secondLock := newDirLock(dir)
	if err := secondLock.tryLock(); err != nil {
		t.Fatalf("unexpected %v", err)
	}
	// must unlock or test can't clean up the tmp dir
	defer secondLock.unlock() //nolint:errcheck

	// retrying original lock should return ErrLocked
	if err := firstLock.tryLock(); !errors.Is(err, errLocked) {
		t.Fatalf("unexpected %v", err)
	}

	// trying to lock a non-existing dir should fails
	if err := newDirLock("/path/to/non/existing/dir").tryLock(); !errors.Is(err, errLockFailed) {
		t.Fatalf("unexpected %v", err)
	}
}

func Test_Lock(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		timeout time.Duration
		expect  error
	}{
		{
			name:    "no timeout",
			timeout: 0,
			expect:  nil,
		},
		{
			name:    "unlocked before timeout",
			timeout: 5 * time.Second,
			expect:  nil,
		},
		{
			name:    "timeout wating for unlock",
			timeout: 1 * time.Second,
			expect:  errLocked,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()

			// get a lock on the tmp dir
			lock := newDirLock(dir)
			err := lock.tryLock()
			if err != nil {
				t.Fatalf("unexpected %v", err)
			}

			// start a goroutine to unlock the file after 3 seconds
			// ensure the file is unlocked after the test
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			t.Cleanup(cancel)
			go func() {
				<-ctx.Done()
				_ = lock.unlock()
			}()

			// try to lock the tmp dir while still locked
			secondLock := newDirLock(dir)
			defer secondLock.unlock() //nolint:errcheck

			err = secondLock.lock(tc.timeout)
			if !errors.Is(err, tc.expect) {
				t.Fatalf("expected %v got %v", tc.expect, err)
			}
		})
	}
}
