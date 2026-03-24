// package main contains everything
package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	tag "github.com/dhowden/tag"
)

var (
	inputDir  = flag.String("i", ".", "the input directory")
	outputDir = flag.String("o", ".", "the output directory")
)

func sanitize(name string) string {
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, ":", "_")
	return strings.TrimSpace(name)
}

func main() {
	ctx := context.Background()
	flag.Parse()

	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		slog.ErrorContext(ctx, "cant create base output dir", slog.String("error", err.Error()))
		os.Exit(1)
	}

	inputFiles, err := os.ReadDir(*inputDir)
	if err != nil {
		slog.ErrorContext(ctx, "cant read input dir", slog.String("error", err.Error()))
		os.Exit(1)
	}

	for _, file := range inputFiles {
		if file.IsDir() {
			continue
		}

		sourcePath := filepath.Join(*inputDir, file.Name())
		f, err := os.Open(sourcePath)
		if err != nil {
			slog.ErrorContext(ctx, "cant open file", slog.String("file", sourcePath), slog.String("error", err.Error()))
			continue
		}

		m, err := tag.ReadFrom(f)

		if fileErr := f.Close(); fileErr != nil {
			slog.ErrorContext(ctx, "cant close file", slog.String("error", fileErr.Error()))
		}

		if err != nil {
			slog.ErrorContext(ctx, "cant read tags (might not be an audio file)", slog.String("file", sourcePath), slog.String("error", err.Error()))
			continue
		}

		if m != nil {
			artist := sanitize(m.Artist())
			album := sanitize(m.Album())

			if artist == "" {
				artist = "Unknown Artist"
			}
			if album == "" {
				album = "Unknown Album"
			}

			destDir := filepath.Join(*outputDir, artist, album)
			if err := os.MkdirAll(destDir, 0755); err != nil {
				slog.ErrorContext(ctx, "cant create artist/album dir", slog.String("dir", destDir), slog.String("error", err.Error()))
				continue
			}

			destPath := filepath.Join(destDir, file.Name())
			if err := os.Rename(sourcePath, destPath); err != nil {
				slog.ErrorContext(ctx, "cant move file", slog.String("source", sourcePath), slog.String("error", err.Error()))
			} else {
				slog.InfoContext(ctx, "moved file successfully",
					slog.String("file", file.Name()),
					slog.String("artist", artist),
					slog.String("album", album))
			}
		}
	}
}
