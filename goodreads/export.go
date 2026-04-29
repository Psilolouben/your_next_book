package goodreads

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

const exportURL = "https://www.goodreads.com/review/import"

func profileDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".goodreads_chrome_profile")
}

func allocator() (context.Context, context.CancelFunc) {
	opts := []chromedp.ExecAllocatorOption{
		chromedp.UserDataDir(profileDir()),
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("headless", false),
	}
	return chromedp.NewExecAllocator(context.Background(), opts...)
}

// FetchExport opens a browser, handles login, waits for the user to click
// Export Library, then clicks the download link and saves the CSV to csvPath.
// Everything happens in one browser session so the auth cookie is never lost.
func FetchExport(csvPath string) error {
	allocCtx, cancel := allocator()
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	if err := chromedp.Run(ctx, chromedp.Navigate(exportURL)); err != nil {
		return fmt.Errorf("failed to open export page: %w", err)
	}

	// Goodreads redirects to sign_in when logged out (including Google OAuth).
	// Wait up to 5 minutes for the user to log in and land back on the export page.
	fmt.Fprintln(os.Stderr, "Waiting to reach export page (log in if prompted)...")
	deadline := time.Now().Add(5 * time.Minute)
	for time.Now().Before(deadline) {
		var u string
		if err := chromedp.Run(ctx, chromedp.Location(&u)); err == nil && strings.HasPrefix(u, exportURL) {
			break
		}
		time.Sleep(time.Second)
	}

	var pageURL string
	chromedp.Run(ctx, chromedp.Location(&pageURL))
	if !strings.HasPrefix(pageURL, exportURL) {
		return fmt.Errorf("login timed out (still at %s)", pageURL)
	}

	// Show a banner. If the export link is already present we skip straight to
	// downloading; otherwise the user needs to click Export Library first.
	chromedp.Run(ctx, chromedp.Evaluate(`
		var b = document.createElement('div');
		b.innerText = '👉 Click "Export Library" to request a fresh export. If the download link is already visible, just click it.';
		b.style = 'position:fixed;top:0;left:0;right:0;background:#ffe;color:#333;font-size:16px;padding:12px 16px;z-index:99999;border-bottom:2px solid #f90;text-align:center;';
		document.body.prepend(b);
	`, nil))

	// Wait up to 5 minutes for the download link to appear.
	fmt.Fprintln(os.Stderr, "Waiting for download link...")
	waitCtx, waitCancel := context.WithTimeout(ctx, 5*time.Minute)
	defer waitCancel()
	if err := chromedp.Run(waitCtx, chromedp.WaitVisible(`a[href*="goodreads_export"]`, chromedp.ByQuery)); err != nil {
		return fmt.Errorf("download link did not appear within 5 minutes")
	}

	// Snapshot ~/Downloads before clicking so we can identify the new file.
	home, _ := os.UserHomeDir()
	downloadsDir := filepath.Join(home, "Downloads")
	before := map[string]struct{}{}
	if entries, _ := os.ReadDir(downloadsDir); entries != nil {
		for _, e := range entries {
			before[e.Name()] = struct{}{}
		}
	}

	if err := chromedp.Run(ctx, chromedp.Click(`a[href*="goodreads_export"]`, chromedp.ByQuery)); err != nil {
		return fmt.Errorf("failed to click download link: %w", err)
	}

	// Wait for the CSV to land in ~/Downloads.
	fmt.Fprintln(os.Stderr, "Waiting for download to complete...")
	deadline = time.Now().Add(30 * time.Second)
	var downloaded string
	for time.Now().Before(deadline) {
		time.Sleep(time.Second)
		entries, _ := os.ReadDir(downloadsDir)
		for _, e := range entries {
			name := e.Name()
			if _, existed := before[name]; existed {
				continue
			}
			if strings.HasSuffix(name, ".crdownload") {
				continue
			}
			if strings.Contains(name, "goodreads") && strings.HasSuffix(name, ".csv") {
				downloaded = filepath.Join(downloadsDir, name)
				break
			}
		}
		if downloaded != "" {
			break
		}
	}

	if downloaded == "" {
		return fmt.Errorf("no goodreads CSV appeared in %s within 30 seconds", downloadsDir)
	}

	if err := os.Rename(downloaded, csvPath); err != nil {
		data, err := os.ReadFile(downloaded)
		if err != nil {
			return fmt.Errorf("failed to read downloaded file: %w", err)
		}
		if err := os.WriteFile(csvPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write CSV: %w", err)
		}
	}

	return nil
}
