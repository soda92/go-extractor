package main

import (
	"errors"
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

var ErrPasswordRequired = errors.New("password required or incorrect password")

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Go Extractor")
	myWindow.Resize(fyne.NewSize(500, 290))

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
	if before, ok := strings.CutSuffix(defaultSubfolder, ".tar"); ok {
		defaultSubfolder = before
	}

	destEntry := widget.NewEntry()
	destEntry.SetText(archiveDir)

	subfolderEntry := widget.NewEntry()
	subfolderEntry.SetText(defaultSubfolder)

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Optional")

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
			{Text: "Password:", Widget: passwordEntry},
		},
	}

	statusLabel := widget.NewLabel("")

	var startExtraction func()
	startExtraction = func() {
		statusLabel.SetText("Extracting...")
		myWindow.Content().Refresh()

		dest := destEntry.Text
		subfolder := subfolderEntry.Text
		password := passwordEntry.Text
		useSubfolder := extractToSubfolder.Checked
		openInDolphin := openInDolphinCheck.Checked

		go func() {
			err := performExtraction(archivePath, dest, subfolder, useSubfolder, password)
			if err != nil {
				if errors.Is(err, ErrPasswordRequired) {
					fyne.Do(func() {
						statusLabel.SetText("Password required...")
						showPasswordPrompt(myWindow, "Encrypted Archive", func(pwd string) {
							passwordEntry.SetText(pwd)
							startExtraction()
						})
					})
				} else {
					fyne.Do(func() {
						dialog.ShowError(err, myWindow)
						statusLabel.SetText("Failed: " + err.Error())
					})
				}
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

	destEntry.OnSubmitted = func(string) { startExtraction() }
	subfolderEntry.OnSubmitted = func(string) { startExtraction() }
	passwordEntry.OnSubmitted = func(string) { startExtraction() }

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

func showPasswordPrompt(myWindow fyne.Window, title string, callback func(string)) {
	pwdEntry := widget.NewPasswordEntry()
	item := widget.NewFormItem("Password:", pwdEntry)
	d := dialog.NewForm(title, "Extract", "Cancel", []*widget.FormItem{item}, func(ok bool) {
		if ok {
			callback(pwdEntry.Text)
		}
	}, myWindow)
	d.Show()
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

func isPasswordError(output string) bool {
	out := strings.ToLower(output)
	return strings.Contains(out, "wrong password") ||
		strings.Contains(out, "encrypted") ||
		strings.Contains(out, "enter password") ||
		strings.Contains(out, "headers error") ||
		strings.Contains(out, "data error")
}

func performExtraction(archivePath, destDir, subfolderName string, useSubfolder bool, password string) error {
	// Create a temporary directory for extraction
	tempDir, err := os.MkdirTemp("", "go-extractor-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Run 7z to extract into the temp dir
	args := []string{"x", "-y", "-o" + tempDir}
	if password != "" {
		args = append(args, "-p"+password)
	} else {
		// -p- disables interactive password prompt on stdin so 7z fails cleanly if password is needed
		args = append(args, "-p-")
	}
	args = append(args, archivePath)

	cmd := exec.Command("7z", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		outStr := string(output)
		if isPasswordError(outStr) {
			return ErrPasswordRequired
		}
		return fmt.Errorf("7z extraction failed: %w\n%s", err, outStr)
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
