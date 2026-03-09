package ociinstaller

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	pfociinstaller "github.com/turbot/pipe-fittings/v2/ociinstaller"
	putils "github.com/turbot/pipe-fittings/v2/utils"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/filepaths"
	versionfile "github.com/turbot/steampipe/pkg/ociinstaller/versionfile"
)

type bundledAssetMetadata struct {
	ImageDigest string `json:"image_digest"`
	ImageRef    string `json:"image_ref"`
	Version     string `json:"version"`
}

func bundledDBAvailable() bool {
	if runtime.GOOS != "linux" {
		return false
	}
	return pfociinstaller.FileExists(constants.EmbeddedDBBundlePath()) && pfociinstaller.FileExists(constants.EmbeddedDBMetadataPath())
}

func bundledFDWAvailable() bool {
	if runtime.GOOS != "linux" {
		return false
	}
	return pfociinstaller.FileExists(constants.EmbeddedFDWMetadataPath()) && pfociinstaller.FileExists(filepath.Join(constants.EmbeddedFDWAssetsDir(), constants.FdwBinaryFileName+".gz"))
}

func BundledDBAvailable() bool {
	return bundledDBAvailable()
}

func BundledFDWAvailable() bool {
	return bundledFDWAvailable()
}

func loadBundledAssetMetadata(path string) (*bundledAssetMetadata, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var metadata bundledAssetMetadata
	if err := json.Unmarshal(content, &metadata); err != nil {
		return nil, err
	}
	if metadata.Version == "" {
		return nil, fmt.Errorf("bundled asset metadata missing version: %s", path)
	}
	if metadata.ImageDigest == "" {
		return nil, fmt.Errorf("bundled asset metadata missing image digest: %s", path)
	}
	return &metadata, nil
}

func writeBundledAssetMetadata(dir string, metadata bundledAssetMetadata) error {
	content, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, constants.EmbeddedAssetMetadataFile), content, 0644)
}

func StageBundledAssets(ctx context.Context, targetOS string, targetArch string, outputRoot string) error {
	if targetOS != "linux" {
		return nil
	}

	if err := os.RemoveAll(outputRoot); err != nil {
		return err
	}
	if err := os.MkdirAll(outputRoot, 0755); err != nil {
		return err
	}

	provider := SteampipeMediaTypeProvider{
		TargetOS:   targetOS,
		TargetArch: targetArch,
	}

	if err := stageBundledDB(ctx, provider, filepath.Join(outputRoot, "db")); err != nil {
		return err
	}
	if err := stageBundledFDW(ctx, provider, filepath.Join(outputRoot, "fdw")); err != nil {
		return err
	}
	return nil
}

func stageBundledDB(ctx context.Context, provider SteampipeMediaTypeProvider, outputDir string) error {
	tempDir := pfociinstaller.NewTempDir(outputDir)
	defer tempDir.Delete()

	imageDownloader := newDbDownloaderForProvider(provider)
	image, err := imageDownloader.Download(ctx, pfociinstaller.NewImageRef(constants.PostgresImageRef), ImageTypeDatabase, tempDir.Path)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}
	if err := CopyFolder(filepath.Join(tempDir.Path, image.Data.ArchiveDir), filepath.Join(outputDir, "postgres")); err != nil {
		return err
	}

	return writeBundledAssetMetadata(outputDir, bundledAssetMetadata{
		ImageDigest: string(image.OCIDescriptor.Digest),
		ImageRef:    image.ImageRef.RequestedRef,
		Version:     image.Config.Database.Version,
	})
}

func stageBundledFDW(ctx context.Context, provider SteampipeMediaTypeProvider, outputDir string) error {
	tempDir := pfociinstaller.NewTempDir(outputDir)
	defer tempDir.Delete()

	imageDownloader := newFdwDownloaderForProvider(provider)
	image, err := imageDownloader.Download(ctx, pfociinstaller.NewImageRef(constants.FdwImageRef), ImageTypeFdw, tempDir.Path)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	if err := CopyFile(filepath.Join(tempDir.Path, image.Data.BinaryFile), filepath.Join(outputDir, constants.FdwBinaryFileName+".gz")); err != nil {
		return err
	}
	if err := CopyFile(filepath.Join(tempDir.Path, image.Data.ControlFile), filepath.Join(outputDir, "steampipe_postgres_fdw.control")); err != nil {
		return err
	}
	if err := CopyFile(filepath.Join(tempDir.Path, image.Data.SqlFile), filepath.Join(outputDir, "steampipe_postgres_fdw--1.0.sql")); err != nil {
		return err
	}

	return writeBundledAssetMetadata(outputDir, bundledAssetMetadata{
		ImageDigest: string(image.OCIDescriptor.Digest),
		ImageRef:    image.ImageRef.RequestedRef,
		Version:     image.Config.Fdw.Version,
	})
}

func InstallDBFromBundle(bundleDir string, dblocation string) (string, error) {
	metadata, err := loadBundledAssetMetadata(filepath.Join(bundleDir, constants.EmbeddedAssetMetadataFile))
	if err != nil {
		return "", err
	}

	source := filepath.Join(bundleDir, "postgres")
	if !pfociinstaller.FileExists(source) {
		return "", fmt.Errorf("bundled database assets not found in %s", source)
	}

	if err = CopyFolder(source, dblocation); err != nil {
		return "", err
	}

	if err := updateVersionFileDB(installedImageDetails{
		version:       metadata.Version,
		imageDigest:   metadata.ImageDigest,
		installedFrom: fmt.Sprintf("bundle:%s", bundleDir),
	}); err != nil {
		return metadata.ImageDigest, err
	}

	return metadata.ImageDigest, nil
}

func InstallFdwFromBundle(bundleDir string) (string, error) {
	metadata, err := loadBundledAssetMetadata(filepath.Join(bundleDir, constants.EmbeddedAssetMetadataFile))
	if err != nil {
		return "", err
	}

	sourceDir := filepath.Join(bundleDir, constants.FdwBinaryFileName+".gz")
	if !pfociinstaller.FileExists(sourceDir) {
		return "", fmt.Errorf("bundled FDW binary not found in %s", sourceDir)
	}

	if err = installBundledFdwFiles(bundleDir); err != nil {
		return "", err
	}

	if err := updateVersionFileFdw(installedImageDetails{
		version:       metadata.Version,
		imageDigest:   metadata.ImageDigest,
		installedFrom: fmt.Sprintf("bundle:%s", bundleDir),
	}); err != nil {
		return metadata.ImageDigest, err
	}

	return metadata.ImageDigest, nil
}

func installBundledFdwFiles(bundleDir string) error {
	fdwBinDir := filepaths.GetFDWBinaryDir()
	fdwBinFileSourcePath := filepath.Join(bundleDir, constants.FdwBinaryFileName+".gz")
	fdwBinFileDestPath := filepath.Join(fdwBinDir, constants.FdwBinaryFileName)

	// Remove the existing binary before replacing it to match the OCI install path.
	os.Remove(fdwBinFileDestPath)
	if _, err := pfociinstaller.Ungzip(fdwBinFileSourcePath, fdwBinDir); err != nil {
		return fmt.Errorf("could not unzip %s to %s: %s", fdwBinFileSourcePath, fdwBinDir, err.Error())
	}

	fdwControlDir := filepaths.GetFDWSQLAndControlDir()
	controlFileSourcePath := filepath.Join(bundleDir, "steampipe_postgres_fdw.control")
	controlFileDestPath := filepath.Join(fdwControlDir, "steampipe_postgres_fdw.control")
	if err := CopyFile(controlFileSourcePath, controlFileDestPath); err != nil {
		return fmt.Errorf("could not install %s to %s", controlFileSourcePath, fdwControlDir)
	}

	sqlFileSourcePath := filepath.Join(bundleDir, "steampipe_postgres_fdw--1.0.sql")
	sqlFileDestPath := filepath.Join(fdwControlDir, "steampipe_postgres_fdw--1.0.sql")
	if err := CopyFile(sqlFileSourcePath, sqlFileDestPath); err != nil {
		return fmt.Errorf("could not install %s to %s", sqlFileSourcePath, fdwControlDir)
	}

	return nil
}

type installedImageDetails struct {
	version       string
	imageDigest   string
	installedFrom string
}

func updateVersionFileDB(details installedImageDetails) error {
	timeNow := putils.FormatTime(time.Now())
	v, err := versionfile.LoadDatabaseVersionFile()
	if err != nil {
		return err
	}
	v.EmbeddedDB.Version = details.version
	v.EmbeddedDB.Name = "embeddedDB"
	v.EmbeddedDB.ImageDigest = details.imageDigest
	v.EmbeddedDB.InstalledFrom = details.installedFrom
	v.EmbeddedDB.LastCheckedDate = timeNow
	v.EmbeddedDB.InstallDate = timeNow
	return v.Save()
}

func updateVersionFileFdw(details installedImageDetails) error {
	timeNow := putils.FormatTime(time.Now())
	v, err := versionfile.LoadDatabaseVersionFile()
	if err != nil {
		return err
	}
	v.FdwExtension.Version = details.version
	v.FdwExtension.Name = "fdwExtension"
	v.FdwExtension.ImageDigest = details.imageDigest
	v.FdwExtension.InstalledFrom = details.installedFrom
	v.FdwExtension.LastCheckedDate = timeNow
	v.FdwExtension.InstallDate = timeNow
	return v.Save()
}
