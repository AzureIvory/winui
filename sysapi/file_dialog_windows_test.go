//go:build windows

package sysapi

import "testing"

func TestNormalizeOptionsDefaultsToOpen(t *testing.T) {
	opts := normalizeOptions(Options{})
	if opts.Mode != DialogOpen {
		t.Fatalf("normalizeOptions() mode = %q, want %q", opts.Mode, DialogOpen)
	}
}

func TestSplitInitialPath(t *testing.T) {
	t.Run("folder only keeps directory", func(t *testing.T) {
		folder, name := splitInitialPath(`C:\work\report.txt`, true)
		if folder != `C:\work` || name != "" {
			t.Fatalf("splitInitialPath(folderOnly) = (%q, %q), want (%q, %q)", folder, name, `C:\work`, "")
		}
	})

	t.Run("file path returns directory and base name", func(t *testing.T) {
		folder, name := splitInitialPath(`C:\work\report.txt`, false)
		if folder != `C:\work` || name != "report.txt" {
			t.Fatalf("splitInitialPath(file) = (%q, %q), want (%q, %q)", folder, name, `C:\work`, "report.txt")
		}
	})

	t.Run("bare file name stays as file name", func(t *testing.T) {
		folder, name := splitInitialPath("report.txt", false)
		if folder != "" || name != "report.txt" {
			t.Fatalf("splitInitialPath(base) = (%q, %q), want (%q, %q)", folder, name, "", "report.txt")
		}
	})
}

func TestNormalizeExtension(t *testing.T) {
	if got := normalizeExtension(".txt"); got != "txt" {
		t.Fatalf("normalizeExtension() = %q, want %q", got, "txt")
	}
}
