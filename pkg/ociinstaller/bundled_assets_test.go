package ociinstaller

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/filepaths"
	versionfile "github.com/turbot/steampipe/pkg/ociinstaller/versionfile"
)

func TestInstallDBFromBundle(t *testing.T) {
	tempDir := t.TempDir()
	previousInstallDir := app_specific.InstallDir
	app_specific.InstallDir = tempDir
	defer func() {
		app_specific.InstallDir = previousInstallDir
	}()

	bundleDir := filepath.Join(tempDir, "bundle-db")
	if err := os.MkdirAll(filepath.Join(bundleDir, "postgres", "bin"), 0755); err != nil {
		t.Fatalf("create bundle dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(bundleDir, "postgres", "bin", "postgres"), []byte("postgres"), 0755); err != nil {
		t.Fatalf("write postgres binary: %v", err)
	}
	writeBundleMetadata(t, bundleDir, bundledAssetMetadata{
		ImageDigest: "sha256:test-db",
		ImageRef:    constants.PostgresImageRef,
		Version:     constants.DatabaseVersion,
	})

	destDir := filepath.Join(tempDir, "installed-db")
	if _, err := InstallDBFromBundle(bundleDir, destDir); err != nil {
		t.Fatalf("install db from bundle: %v", err)
	}

	if _, err := os.Stat(filepath.Join(destDir, "bin", "postgres")); err != nil {
		t.Fatalf("expected bundled postgres binary to be installed: %v", err)
	}

	versionInfo, err := versionfile.LoadDatabaseVersionFile()
	if err != nil {
		t.Fatalf("load version file: %v", err)
	}
	if versionInfo.EmbeddedDB.Version != constants.DatabaseVersion {
		t.Fatalf("unexpected db version %q", versionInfo.EmbeddedDB.Version)
	}
	if versionInfo.EmbeddedDB.ImageDigest != "sha256:test-db" {
		t.Fatalf("unexpected db digest %q", versionInfo.EmbeddedDB.ImageDigest)
	}
	if versionInfo.EmbeddedDB.InstalledFrom != "bundle:"+bundleDir {
		t.Fatalf("unexpected db install source %q", versionInfo.EmbeddedDB.InstalledFrom)
	}
}

func TestInstallFdwFromBundle(t *testing.T) {
	tempDir := t.TempDir()
	previousInstallDir := app_specific.InstallDir
	app_specific.InstallDir = tempDir
	defer func() {
		app_specific.InstallDir = previousInstallDir
	}()

	if err := os.MkdirAll(filepaths.GetFDWBinaryDir(), 0755); err != nil {
		t.Fatalf("create fdw binary dir: %v", err)
	}
	if err := os.MkdirAll(filepaths.GetFDWSQLAndControlDir(), 0755); err != nil {
		t.Fatalf("create fdw control dir: %v", err)
	}

	bundleDir := filepath.Join(tempDir, "bundle-fdw")
	if err := os.MkdirAll(bundleDir, 0755); err != nil {
		t.Fatalf("create bundle dir: %v", err)
	}
	writeBundleMetadata(t, bundleDir, bundledAssetMetadata{
		ImageDigest: "sha256:test-fdw",
		ImageRef:    constants.FdwImageRef,
		Version:     constants.FdwVersion,
	})
	writeGzipFile(t, filepath.Join(bundleDir, constants.FdwBinaryFileName+".gz"), constants.FdwBinaryFileName, []byte("fdw-binary"))
	if err := os.WriteFile(filepath.Join(bundleDir, "steampipe_postgres_fdw.control"), []byte("control"), 0644); err != nil {
		t.Fatalf("write control file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(bundleDir, "steampipe_postgres_fdw--1.0.sql"), []byte("sql"), 0644); err != nil {
		t.Fatalf("write sql file: %v", err)
	}

	if _, err := InstallFdwFromBundle(bundleDir); err != nil {
		t.Fatalf("install fdw from bundle: %v", err)
	}

	binaryPath := filepath.Join(filepaths.GetFDWBinaryDir(), constants.FdwBinaryFileName)
	content, err := os.ReadFile(binaryPath)
	if err != nil {
		t.Fatalf("read installed fdw binary: %v", err)
	}
	if string(content) != "fdw-binary" {
		t.Fatalf("unexpected fdw binary content %q", string(content))
	}

	versionInfo, err := versionfile.LoadDatabaseVersionFile()
	if err != nil {
		t.Fatalf("load version file: %v", err)
	}
	if versionInfo.FdwExtension.Version != constants.FdwVersion {
		t.Fatalf("unexpected fdw version %q", versionInfo.FdwExtension.Version)
	}
	if versionInfo.FdwExtension.ImageDigest != "sha256:test-fdw" {
		t.Fatalf("unexpected fdw digest %q", versionInfo.FdwExtension.ImageDigest)
	}
	if versionInfo.FdwExtension.InstalledFrom != "bundle:"+bundleDir {
		t.Fatalf("unexpected fdw install source %q", versionInfo.FdwExtension.InstalledFrom)
	}
}

func writeBundleMetadata(t *testing.T, bundleDir string, metadata bundledAssetMetadata) {
	t.Helper()

	content, err := json.Marshal(metadata)
	if err != nil {
		t.Fatalf("marshal metadata: %v", err)
	}
	if err := os.WriteFile(filepath.Join(bundleDir, constants.EmbeddedAssetMetadataFile), content, 0644); err != nil {
		t.Fatalf("write metadata: %v", err)
	}
}

func writeGzipFile(t *testing.T, path string, name string, content []byte) {
	t.Helper()

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create gzip file: %v", err)
	}
	defer file.Close()

	writer := gzip.NewWriter(file)
	writer.Name = name
	if _, err := writer.Write(content); err != nil {
		t.Fatalf("write gzip content: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close gzip writer: %v", err)
	}
}
