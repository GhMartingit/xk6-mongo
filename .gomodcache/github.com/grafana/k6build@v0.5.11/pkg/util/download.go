package util

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"syscall"
)

var (
	ErrDownloadFailed = fmt.Errorf("downloading file failed")    //nolint:revive
	ErrWritingFile    = fmt.Errorf("opening output file failed") //nolint:revive
)

// Download downloads a file from a URL and saves it to the output file.
func Download(ctx context.Context, url string, output string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("%w %w", ErrDownloadFailed, err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w %w", ErrDownloadFailed, err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w %w", ErrDownloadFailed, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	outFile, err := os.OpenFile( //nolint:gosec
		output,
		os.O_TRUNC|os.O_WRONLY|os.O_CREATE,
		syscall.S_IRUSR|syscall.S_IXUSR|syscall.S_IWUSR,
	)
	if err != nil {
		return fmt.Errorf("%w %w", ErrWritingFile, err)
	}
	defer func() {
		_ = outFile.Close()
	}()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("%w %w", ErrWritingFile, err)
	}

	return nil
}
