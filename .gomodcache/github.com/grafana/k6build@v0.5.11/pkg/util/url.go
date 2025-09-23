// Package util implements utility functions
package util

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"
)

// Adapted from https://go-review.googlesource.com/c/vuln/+/438175/7/internal/web/url.go

// URLFromFilePath converts the given absolute path to a URL.
func URLFromFilePath(path string) (*url.URL, error) {
	if !filepath.IsAbs(path) {
		return nil, fmt.Errorf("path is not absolute %q", path)
	}

	// If path has a Windows volume name, convert the volume to a host and prefix
	// per https://blogs.msdn.microsoft.com/ie/2006/12/06/file-uris-in-windows/.
	if vol := filepath.VolumeName(path); vol != "" {
		if strings.HasPrefix(vol, `\\`) {
			path = filepath.ToSlash(path[2:])
			i := strings.IndexByte(path, '/')

			if i < 0 {
				// A degenerate case.
				// \\host.example.com (without a share name)
				// becomes
				// file://host.example.com/
				return &url.URL{
					Scheme: "file",
					Host:   path,
					Path:   "/",
				}, nil
			}

			// \\host.example.com\Share\path\to\file
			// becomes
			// file://host.example.com/Share/path/to/file
			return &url.URL{
				Scheme: "file",
				Host:   path[:i],
				Path:   filepath.ToSlash(path[i:]),
			}, nil
		}

		// C:\path\to\file
		// becomes
		// file:///C:/path/to/file
		return &url.URL{
			Scheme: "file",
			Path:   "/" + filepath.ToSlash(path),
		}, nil
	}

	// /path/to/file
	// becomes
	// file:///path/to/file
	return &url.URL{
		Scheme: "file",
		Path:   filepath.ToSlash(path),
	}, nil
}

// URLToFilePath converts a file-scheme url to a file path.
func URLToFilePath(u *url.URL) (string, error) {
	if u.Scheme != "file" {
		return "", errors.New("non-file URL")
	}

	checkAbs := func(path string) (string, error) {
		if !filepath.IsAbs(path) {
			return "", fmt.Errorf("path is not absolute %q", path)
		}
		return path, nil
	}

	if u.Path == "" {
		if u.Host != "" || u.Opaque == "" {
			return "", errors.New("file URL missing path")
		}
		return checkAbs(filepath.FromSlash(u.Opaque))
	}

	path, err := convertFileURLPath(u.Host, u.Path)
	if err != nil {
		return path, err
	}
	return checkAbs(path)
}

func convertFileURLPath(host, path string) (string, error) {
	if runtime.GOOS == "windows" {
		return convertFileURLPathWindows(host, path)
	}
	switch host {
	case "", "localhost":
	default:
		return "", errors.New("file URL specifies non-local host")
	}
	return filepath.FromSlash(path), nil
}

func convertFileURLPathWindows(host, path string) (string, error) {
	if len(path) == 0 || path[0] != '/' {
		return "", fmt.Errorf("path is not absolute %q", path)
	}

	path = filepath.FromSlash(path)

	// We interpret Windows file URLs per the description in
	// https://blogs.msdn.microsoft.com/ie/2006/12/06/file-uris-in-windows/.

	// The host part of a file URL (if any) is the UNC volume name,
	// but RFC 8089 reserves the authority "localhost" for the local machine.
	if host != "" && host != "localhost" {
		// A common "legacy" format omits the leading slash before a drive letter,
		// encoding the drive letter as the host instead of part of the path.
		// (See https://blogs.msdn.microsoft.com/freeassociations/2005/05/19/the-bizarre-and-unhappy-story-of-file-urls/.)
		// We do not support that format, but we should at least emit a more
		// helpful error message for it.
		if filepath.VolumeName(host) != "" {
			return "", errors.New("file URL encodes volume in host field: too few slashes?")
		}
		return `\\` + host + path, nil
	}

	// If host is empty, path must contain an initial slash followed by a
	// drive letter and path. Remove the slash and verify that the path is valid.
	if vol := filepath.VolumeName(path[1:]); vol == "" || strings.HasPrefix(vol, `\\`) {
		return "", errors.New("file URL missing drive letter")
	}
	return path[1:], nil
}
