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

	// Inject a banner we can update throughout the flow.
	chromedp.Run(ctx, chromedp.Evaluate(`
		var b = document.createElement('div');
		b.id = 'gr-export-banner';
		b.style = 'position:fixed;top:0;left:0;right:0;background:#ffe;color:#333;font-size:16px;padding:12px 16px;z-index:99999;border-bottom:2px solid #f90;text-align:center;';
		document.body.prepend(b);
	`, nil))

	chromedp.Run(ctx, chromedp.Evaluate(`document.getElementById('gr-export-banner').innerText = '👉 Click "Export Library" to generate a fresh export, then wait here.';`, nil))
	fmt.Fprintln(os.Stderr, "Waiting for user to click Export Library...")

	// Give the page's JS up to 5 seconds to render any existing export link.
	// If one appears, it's stale — wait for it to disappear (user clicked Export),
	// then wait for the fresh one. If nothing appears within 5s, the page is clean
	// and we just wait for the new link.
	probeCtx, probeCancel := context.WithTimeout(ctx, 5*time.Second)
	defer probeCancel()
	existingLinkFound := chromedp.Run(probeCtx, chromedp.WaitVisible(`a[href*="goodreads_export"]`, chromedp.ByQuery)) == nil

	if existingLinkFound {
		fmt.Fprintln(os.Stderr, "Existing export link found — waiting for user to replace it...")
		staleCtx, staleCancel := context.WithTimeout(ctx, 5*time.Minute)
		defer staleCancel()
		chromedp.Run(staleCtx, chromedp.WaitNotPresent(`a[href*="goodreads_export"]`, chromedp.ByQuery))
	}

	chromedp.Run(ctx, chromedp.Evaluate(`document.getElementById('gr-export-banner').innerText = '⏳ Waiting for your export to be ready...';`, nil))

	// Wait up to 5 minutes for the fresh download link to appear.
	fmt.Fprintln(os.Stderr, "Waiting for fresh download link...")
	waitCtx, waitCancel := context.WithTimeout(ctx, 5*time.Minute)
	defer waitCancel()
	if err := chromedp.Run(waitCtx, chromedp.WaitVisible(`a[href*="goodreads_export"]`, chromedp.ByQuery)); err != nil {
		return fmt.Errorf("download link did not appear within 5 minutes")
	}
	chromedp.Run(ctx, chromedp.Evaluate(`document.getElementById('gr-export-banner').innerText = '⬇️ Downloading your export...';`, nil))

	// Snapshot ~/Downloads before clicking so we can identify the new file.
	home, _ := os.UserHomeDir()
	downloadsDir := filepath.Join(home, "Downloads")
	before := map[string]struct{}{}
	if entries, _ := os.ReadDir(downloadsDir); entries != nil {
		for _, e := range entries {
			before[e.Name()] = struct{}{}
		}
	}

	// Get the href and fetch it directly via the browser to avoid navigation issues.
	var downloadHref string
	if err := chromedp.Run(ctx, chromedp.Evaluate(`(function(){ var a = document.querySelector('a[href*="goodreads_export"]'); return a ? a.href : ''; })()`, &downloadHref)); err != nil || downloadHref == "" {
		return fmt.Errorf("could not read download link href")
	}
	fmt.Fprintf(os.Stderr, "Download URL: %s\n", downloadHref)

	// Navigate to the download URL directly — more reliable than clicking.
	if err := chromedp.Run(ctx, chromedp.Navigate(downloadHref)); err != nil {
		// Navigation "fails" when the browser downloads instead of navigating — this is expected.
		fmt.Fprintf(os.Stderr, "Navigate returned (expected for file downloads): %v\n", err)
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
