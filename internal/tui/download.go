package tui

import (
	"fmt"

	"github.com/chr-hen/mtg-tui/internal/api"
	"github.com/rivo/tview"
)

func (a *App) downloadBulkData() {
	// Show download modal - make it visible immediately
	modal := tview.NewModal().
		SetText("Downloading card database from Scryfall...\nThis may take a few minutes.\n\nPlease wait...").
		AddButtons([]string{})

	a.app.QueueUpdate(func() {
		a.pages.AddPage("download", modal, true, true)
		a.pages.SwitchToPage("download")
	})

	// Fetch bulk data metadata
	bulkDataList, err := api.GetBulkDataMetadata()
	if err != nil {
		a.app.QueueUpdate(func() {
			errorModal := tview.NewModal().
				SetText(fmt.Sprintf("Error: %v", err)).
				AddButtons([]string{"OK"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					a.pages.RemovePage("error")
					a.app.Stop()
				})
			a.pages.AddPage("error", errorModal, true, true)
		})
		return
	}

	// Find default_cards
	defaultCards, err := api.GetDefaultCardsBulkData(bulkDataList)
	if err != nil {
		a.app.QueueUpdate(func() {
			errorModal := tview.NewModal().
				SetText(fmt.Sprintf("Error: %v", err)).
				AddButtons([]string{"OK"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					a.pages.RemovePage("error")
					a.app.Stop()
				})
			a.pages.AddPage("error", errorModal, true, true)
		})
		return
	}

	// Update modal with file size info and download location
	sizeMB := float64(defaultCards.Size) / (1024 * 1024)
	cachePath := api.GetCacheFilePath()
	a.app.QueueUpdate(func() {
		modal.SetText(fmt.Sprintf("Downloading card database...\n\nFile size: %.1f MB\nSaving to: %s\n\nPlease wait...", sizeMB, cachePath))
	})

	// Download with progress updates
	err = api.DownloadBulkData(defaultCards.DownloadURI, func(bytesRead, totalBytes int64) {
		downloadedMB := float64(bytesRead) / (1024 * 1024)
		var progressText string

		if totalBytes > 0 {
			percent := float64(bytesRead) / float64(totalBytes) * 100
			totalMB := float64(totalBytes) / (1024 * 1024)
			progressText = fmt.Sprintf("Downloading card database...\n\nProgress: %.1f%%\n%.1f MB / %.1f MB", percent, downloadedMB, totalMB)
		} else {
			// Content-Length not available, just show bytes downloaded
			progressText = fmt.Sprintf("Downloading card database...\n\nDownloaded: %.1f MB\n\nPlease wait...", downloadedMB)
		}

		a.app.QueueUpdate(func() {
			modal.SetText(progressText)
		})
	})

	if err != nil {
		a.app.QueueUpdate(func() {
			errorModal := tview.NewModal().
				SetText(fmt.Sprintf("Error downloading: %v", err)).
				AddButtons([]string{"OK"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					a.pages.RemovePage("error")
					a.app.Stop()
				})
			a.pages.AddPage("error", errorModal, true, true)
		})
		return
	}

	// Remove download modal
	a.app.QueueUpdate(func() {
		a.pages.RemovePage("download")
	})
}

