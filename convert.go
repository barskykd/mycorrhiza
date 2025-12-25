package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/bouncepaw/mycorrhiza/history"
	"github.com/bouncepaw/mycorrhiza/internal/converter"
	"github.com/bouncepaw/mycorrhiza/internal/files"
	"github.com/bouncepaw/mycorrhiza/internal/hyphae"
)

// convertFormatCommand converts all hyphae to the specified format
func convertFormatCommand(targetFormat string) error {
	if err := files.PrepareWikiRoot(); err != nil {
		slog.Error("Failed to prepare wiki root", "err", err)
		return err
	}

	if err := os.Chdir(files.HyphaeDir()); err != nil {
		slog.Error("Failed to chdir to hyphae dir",
			"err", err, "hyphaeDir", files.HyphaeDir())
		return err
	}

	// Initialize git
	if err := history.Start(); err != nil {
		return err
	}
	history.InitGitRepo()

	// Index all hyphae
	slog.Info("Indexing hyphae...")
	hyphae.Index(files.HyphaeDir())

	// Parse target format
	var toFormat hyphae.TextFormat
	switch strings.ToLower(targetFormat) {
	case "markdown", "md":
		toFormat = hyphae.FormatMarkdown
	case "mycomarkup", "myco":
		toFormat = hyphae.FormatMycomarkup
	default:
		return fmt.Errorf("unknown format: %s (use 'markdown' or 'mycomarkup')", targetFormat)
	}

	// Count hyphae that will be converted
	var needsConversion int
	for h := range hyphae.FilterHyphaeWithText(hyphae.YieldExistingHyphae()) {
		if hyphae.DetectTextFormat(h.TextFilePath()) != toFormat {
			needsConversion++
		}
	}

	if needsConversion == 0 {
		slog.Info("All hyphae are already in the target format", "format", hyphae.FormatName(toFormat))
		return nil
	}

	// Confirm conversion
	fmt.Printf("\nWARNING: This will convert %d hyphae to %s format.\n", needsConversion, hyphae.FormatName(toFormat))
	fmt.Printf("This operation will modify files in your wiki directory.\n")
	fmt.Printf("It is STRONGLY recommended to backup your wiki before proceeding.\n\n")
	fmt.Printf("Continue? (yes/no): ")

	var response string
	fmt.Scanln(&response)
	if strings.ToLower(strings.TrimSpace(response)) != "yes" {
		slog.Info("Conversion cancelled by user")
		return nil
	}

	slog.Info("Starting format conversion", "targetFormat", hyphae.FormatName(toFormat))

	// Get all hyphae with text
	allHyphae := hyphae.FilterHyphaeWithText(hyphae.YieldExistingHyphae())

	var (
		converted     int
		skipped       int
		failed        int
		renamedFiles  = make(map[string]string) // old path -> new path
		modifiedFiles []string                   // files that were modified in place
	)

	for h := range allHyphae {
		oldPath := h.TextFilePath()
		fromFormat := hyphae.DetectTextFormat(oldPath)

		// Skip if already in target format
		if fromFormat == toFormat {
			skipped++
			continue
		}

		// Read content
		content, err := os.ReadFile(oldPath)
		if err != nil {
			slog.Error("Failed to read hypha",
				"hypha", h.CanonicalName(),
				"path", oldPath,
				"err", err)
			failed++
			continue
		}

		// Convert content
		convertedContent, warnings, err := converter.ConvertFormat(string(content), fromFormat, toFormat)
		if err != nil {
			slog.Error("Failed to convert hypha",
				"hypha", h.CanonicalName(),
				"err", err)
			failed++
			continue
		}

		// Log warnings
		if len(warnings) > 0 {
			for _, warning := range warnings {
				slog.Warn("Conversion warning",
					"hypha", h.CanonicalName(),
					"warning", warning)
			}
		}

		// Calculate new path
		newPath := replaceExtension(oldPath, hyphae.FormatExtension(toFormat))

		// Write converted content to new path
		if err := os.WriteFile(newPath, []byte(convertedContent), 0666); err != nil {
			slog.Error("Failed to write converted hypha",
				"hypha", h.CanonicalName(),
				"path", newPath,
				"err", err)
			failed++
			continue
		}

		// Track file changes for git commit
		if oldPath != newPath {
			// File was renamed (extension changed)
			renamedFiles[oldPath] = newPath
		} else {
			// File was modified in place (same extension)
			modifiedFiles = append(modifiedFiles, newPath)
		}

		// Remove old file if it has a different path
		if oldPath != newPath {
			if err := os.Remove(oldPath); err != nil {
				slog.Warn("Failed to remove old hypha file",
					"hypha", h.CanonicalName(),
					"path", oldPath,
					"err", err)
				// Don't count as failed - the conversion succeeded
			}
		}

		// Update hypha in storage
		hyphae.RenameHyphaTo(h, h.CanonicalName(), func(path string) string {
			return replaceExtension(path, hyphae.FormatExtension(toFormat))
		})

		converted++
		if converted%100 == 0 {
			slog.Info("Conversion progress",
				"converted", converted,
				"skipped", skipped,
				"failed", failed)
		}
	}

	slog.Info("Conversion complete",
		"converted", converted,
		"skipped", skipped,
		"failed", failed,
		"total", converted+skipped+failed)

	if failed > 0 {
		return fmt.Errorf("conversion completed with %d failures", failed)
	}

	// Commit changes to git
	if converted > 0 {
		slog.Info("Committing changes to git repository...")
		commitMsg := fmt.Sprintf("Convert %d hyphae to %s format", converted, hyphae.FormatName(toFormat))

		hop := history.Operation(history.TypeMarkupMigration).
			WithMsg(commitMsg)

		// For renamed files: remove old, add new
		// (we've already done the file system operations)
		if len(renamedFiles) > 0 {
			var oldPaths []string
			var newPaths []string
			for oldPath, newPath := range renamedFiles {
				oldPaths = append(oldPaths, oldPath)
				newPaths = append(newPaths, newPath)
			}
			hop.WithFilesRemoved(oldPaths...)
			hop.WithFiles(newPaths...)
		}

		// Add modified files (git add)
		if len(modifiedFiles) > 0 {
			hop.WithFiles(modifiedFiles...)
		}

		// Apply the commit
		hop.Apply()

		if hop.HasErrors() {
			slog.Error("Git commit failed", "err", hop.FirstErrorText())
			return fmt.Errorf("failed to commit changes: %s", hop.FirstErrorText())
		}

		slog.Info("Changes committed to git repository")
	}

	return nil
}

// replaceExtension replaces the file extension in a path
func replaceExtension(path, newExt string) string {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	return filepath.Join(dir, name+newExt)
}
