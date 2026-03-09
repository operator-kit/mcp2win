package selfupdate

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestArchiveName(t *testing.T) {
	got := archiveName("0.2.0")
	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}
	want := fmt.Sprintf("mcp2win_0.2.0_%s_%s.%s", runtime.GOOS, runtime.GOARCH, ext)
	if got != want {
		t.Errorf("archiveName = %q, want %q", got, want)
	}
}

func TestFindChecksum(t *testing.T) {
	checksums := []byte("abc123  file_a.tar.gz\ndef456  file_b.zip\n")

	t.Run("found", func(t *testing.T) {
		got, err := findChecksum(checksums, "file_a.tar.gz")
		if err != nil {
			t.Fatal(err)
		}
		if got != "abc123" {
			t.Errorf("got %q, want %q", got, "abc123")
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := findChecksum(checksums, "nonexistent.tar.gz")
		if err == nil {
			t.Fatal("expected error for missing checksum")
		}
	})
}

func TestVerifyChecksum(t *testing.T) {
	data := []byte("hello world")
	h := sha256.Sum256(data)
	hash := hex.EncodeToString(h[:])

	t.Run("match", func(t *testing.T) {
		if err := verifyChecksum(data, hash); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("mismatch", func(t *testing.T) {
		if err := verifyChecksum(data, "wrong"); err == nil {
			t.Fatal("expected error for mismatched checksum")
		}
	})
}

func createTestTarGz(t *testing.T, files map[string][]byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	for name, content := range files {
		hdr := &tar.Header{
			Name: name,
			Size: int64(len(content)),
			Mode: 0o755,
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write(content); err != nil {
			t.Fatal(err)
		}
	}
	tw.Close()
	gw.Close()
	return buf.Bytes()
}

func createTestZip(t *testing.T, files map[string][]byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, content := range files {
		fw, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fw.Write(content); err != nil {
			t.Fatal(err)
		}
	}
	zw.Close()
	return buf.Bytes()
}

func TestExtractTarGz(t *testing.T) {
	data := createTestTarGz(t, map[string][]byte{
		"mcp2win": []byte("binary-content"),
	})
	dir := t.TempDir()
	if err := extractTarGz(data, dir); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filepath.Join(dir, "mcp2win"))
	if err != nil {
		t.Fatalf("read mcp2win: %v", err)
	}
	if string(got) != "binary-content" {
		t.Errorf("content = %q, want %q", got, "binary-content")
	}
}

func TestExtractZip(t *testing.T) {
	data := createTestZip(t, map[string][]byte{
		"mcp2win.exe": []byte("binary-content"),
	})
	dir := t.TempDir()
	if err := extractZip(data, dir); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filepath.Join(dir, "mcp2win.exe"))
	if err != nil {
		t.Fatalf("read mcp2win.exe: %v", err)
	}
	if string(got) != "binary-content" {
		t.Errorf("content = %q, want %q", got, "binary-content")
	}
}

func TestReplaceBinary(t *testing.T) {
	dir := t.TempDir()

	dst := filepath.Join(dir, "mcp2win")
	if err := os.WriteFile(dst, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}

	src := filepath.Join(dir, "mcp2win-new")
	if err := os.WriteFile(src, []byte("new"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := replaceBinary(src, dst); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new" {
		t.Errorf("binary content = %q, want %q", got, "new")
	}
}

func TestUpdate(t *testing.T) {
	binName := "mcp2win"
	if runtime.GOOS == "windows" {
		binName = "mcp2win.exe"
	}

	var archiveData []byte
	if runtime.GOOS == "windows" {
		archiveData = createTestZip(t, map[string][]byte{
			binName: []byte("new-mcp2win"),
		})
	} else {
		archiveData = createTestTarGz(t, map[string][]byte{
			binName: []byte("new-mcp2win"),
		})
	}

	h := sha256.Sum256(archiveData)
	hash := hex.EncodeToString(h[:])
	archive := archiveName("0.2.0")
	checksumData := fmt.Sprintf("%s  %s\n", hash, archive)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/checksums.txt":
			w.Write([]byte(checksumData))
		case "/archive":
			w.Write(archiveData)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	installDir := t.TempDir()
	os.WriteFile(filepath.Join(installDir, binName), []byte("old-mcp2win"), 0o755)

	origInstallDir := InstallDirOverride
	InstallDirOverride = installDir
	t.Cleanup(func() { InstallDirOverride = origInstallDir })

	release := &ReleaseResponse{
		TagName: "v0.2.0",
		Assets: []Asset{
			{Name: archive, BrowserDownloadURL: srv.URL + "/archive"},
			{Name: "checksums.txt", BrowserDownloadURL: srv.URL + "/checksums.txt"},
		},
	}

	var buf bytes.Buffer
	if err := Update(release, &buf); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(installDir, binName))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new-mcp2win" {
		t.Errorf("mcp2win content = %q, want %q", got, "new-mcp2win")
	}
}

func TestIsNpmInstall(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"/usr/local/bin/mcp2win", false},
		{"/home/user/.local/bin/mcp2win", false},
		{"/usr/local/lib/node_modules/@operatorkit/mcp2win/bin/mcp2win", true},
		{"C:\\Users\\me\\AppData\\Roaming\\npm\\mcp2win.exe", true},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := isNpmInstall(tt.path)
			if got != tt.want {
				t.Errorf("isNpmInstall(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}
