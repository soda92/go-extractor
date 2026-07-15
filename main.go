package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Go Extractor")
	myWindow.Resize(fyne.NewSize(500, 250))

	if len(os.Args) < 2 {
		showErrorAndExit(myApp, myWindow, "No archive file specified.\nUsage: go-extractor <archive-path>")
		return
	}

	archivePath, err := filepath.Abs(os.Args[1])
	if err != nil {
		showErrorAndExit(myApp, myWindow, fmt.Sprintf("Invalid path: %v", err))
		return
	}

	archiveName := filepath.Base(archivePath)
	archiveDir := filepath.Dir(archivePath)

	// Default subfolder name is the archive name without suffix.
	ext := filepath.Ext(archiveName)
	defaultSubfolder := strings.TrimSuffix(archiveName, ext)
	if before, ok :=strings.CutSuffix(defaultSubfolder, ".tar"); ok  {
		defaultSubfolder = before
	}

	destEntry := widget.NewEntry()
	destEntry.SetText(archiveDir)

	subfolderEntry := widget.NewEntry()
	subfolderEntry.SetText(defaultSubfolder)

	extractToSubfolder := widget.NewCheck("Extract to subfolder", func(checked bool) {
		if checked {
			subfolderEntry.Enable()
		} else {
			subfolderEntry.Disable()
		}
	})
	extractToSubfolder.SetChecked(true)

	openInDolphinCheck := widget.NewCheck("Open in Dolphin", nil)
	openInDolphinCheck.SetChecked(true)

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Archive:", Widget: widget.NewLabel(archiveName)},
			{Text: "Destination:", Widget: destEntry},
			{Text: "Subfolder Name:", Widget: subfolderEntry},
		},
	}

	statusLabel := widget.NewLabel("")

	var startExtraction func()
	startExtraction = func() {
		statusLabel.SetText("Extracting...")
		myWindow.Content().Refresh()

		dest := destEntry.Text
		subfolder := subfolderEntry.Text
		useSubfolder := extractToSubfolder.Checked
		openInDolphin := openInDolphinCheck.Checked

		go func() {
			err := performExtraction(archivePath, dest, subfolder, useSubfolder)
			if err != nil {
				fyne.Do(func() {
					dialog.ShowError(err, myWindow)
					statusLabel.SetText("Failed: " + err.Error())
				})
			} else {
				if openInDolphin {
					var targetFolder string
					if useSubfolder {
						targetFolder = filepath.Join(dest, subfolder)
					} else {
						targetFolder = dest
					}
					cmd := exec.Command("dolphin", targetFolder)
					_ = cmd.Start()
				}
				fyne.Do(func() {
					dialog.ShowInformation("Success", "Archive successfully extracted!", myWindow)
					myApp.Quit()
				})
			}
		}()
	}

	destEntry.OnSubmitted = func(string) {
		startExtraction()
	}
	subfolderEntry.OnSubmitted = func(string) {
		startExtraction()
	}

	extractBtn := widget.NewButton("Extract", startExtraction)

	cancelBtn := widget.NewButton("Cancel", func() {
		myApp.Quit()
	})

	buttons := container.NewHBox(extractBtn, cancelBtn)
	content := container.NewVBox(
		form,
		extractToSubfolder,
		openInDolphinCheck,
		statusLabel,
		buttons,
	)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}

func showErrorAndExit(myApp fyne.App, myWindow fyne.Window, msg string) {
	d := dialog.NewError(fmt.Errorf("%s", msg), myWindow)
	d.SetOnClosed(func() {
		myApp.Quit()
	})
	d.Show()
	myWindow.ShowAndRun()
}

func movePath(src, dst string) error {
	// Try standard rename first
	err := os.Rename(src, dst)
	if err == nil {
		return nil
	}
	// Fall back to system mv command if rename fails (e.g., cross-device boundary)
	cmd := exec.Command("mv", src, dst)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to move %s to %s: %w", src, dst, err)
	}
	return nil
}

func performExtraction(archivePath, destDir, subfolderName string, useSubfolder bool) error {
	// Create a temporary directory for extraction
	tempDir, err := os.MkdirTemp("", "go-extractor-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Run 7z to extract into the temp dir
	cmd := exec.Command("7z", "x", "-o"+tempDir, archivePath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("7z extraction failed: %w", err)
	}

	// Read the temp directory contents
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		return fmt.Errorf("failed to read temp dir: %w", err)
	}

	if len(entries) == 0 {
		return fmt.Errorf("archive is empty")
	}

	// Detect if it is a single-folder archive
	isSingleFolder := false
	var singleFolderName string
	if len(entries) == 1 && entries[0].IsDir() {
		isSingleFolder = true
		singleFolderName = entries[0].Name()
	}

	var finalDest string
	if useSubfolder {
		finalDest = filepath.Join(destDir, subfolderName)
	} else {
		finalDest = destDir
	}

	// Create the final destination directory if we are using a subfolder
	if useSubfolder {
		if err := os.MkdirAll(finalDest, 0755); err != nil {
			return fmt.Errorf("failed to create destination: %w", err)
		}
	}

	// Determine source items and move them to finalDest
	if isSingleFolder {
		// Single-folder archive
		srcDir := filepath.Join(tempDir, singleFolderName)
		items, err := os.ReadDir(srcDir)
		if err != nil {
			return fmt.Errorf("failed to read single folder: %w", err)
		}

		if useSubfolder {
			// Option 2: rename/flatten. Move contents of srcDir directly to finalDest
			for _, item := range items {
				srcPath := filepath.Join(srcDir, item.Name())
				dstPath := filepath.Join(finalDest, item.Name())
				if err := movePath(srcPath, dstPath); err != nil {
					return fmt.Errorf("failed to move entry: %w", err)
				}
			}
		} else {
			// Extract single folder directly to destDir (move the single folder itself)
			dstPath := filepath.Join(destDir, singleFolderName)
			if err := movePath(srcDir, dstPath); err != nil {
				return fmt.Errorf("failed to move folder: %w", err)
			}
		}
	} else {
		// Multi-file/folder archive
		for _, item := range entries {
			srcPath := filepath.Join(tempDir, item.Name())
			dstPath := filepath.Join(finalDest, item.Name())
			if err := movePath(srcPath, dstPath); err != nil {
				return fmt.Errorf("failed to move entry: %w", err)
			}
		}
	}

	return nil
}
