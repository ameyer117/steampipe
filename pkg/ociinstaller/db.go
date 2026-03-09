package ociinstaller

import (
	"context"
	"log"
	"path/filepath"

	"github.com/turbot/pipe-fittings/v2/ociinstaller"
	"github.com/turbot/steampipe/pkg/constants"
)

// InstallDB :: Install Postgres files fom OCI image
func InstallDB(ctx context.Context, dblocation string) (string, error) {
	tempDir := ociinstaller.NewTempDir(dblocation)
	defer func() {
		if err := tempDir.Delete(); err != nil {
			log.Printf("[TRACE] Failed to delete temp dir '%s' after installing db files: %s", tempDir, err)
		}
	}()

	imageDownloader := newDbDownloader()

	// Download the blobs
	image, err := imageDownloader.Download(ctx, ociinstaller.NewImageRef(constants.PostgresImageRef), ImageTypeDatabase, tempDir.Path)
	if err != nil {
		return "", err
	}

	// install the files
	if err = installDbFiles(image, tempDir.Path, dblocation); err != nil {
		return "", err
	}

	if err := updateVersionFileDB(installedImageDetails{
		version:       image.Config.Database.Version,
		imageDigest:   string(image.OCIDescriptor.Digest),
		installedFrom: image.ImageRef.RequestedRef,
	}); err != nil {
		return string(image.OCIDescriptor.Digest), err
	}
	return string(image.OCIDescriptor.Digest), nil
}

func installDbFiles(image *ociinstaller.OciImage[*dbImage, *dbImageConfig], tempDir string, dest string) error {
	source := filepath.Join(tempDir, image.Data.ArchiveDir)
	return ociinstaller.MoveFolderWithinPartition(source, dest)
}
