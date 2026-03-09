package constants

import "path/filepath"

const (
	// EmbeddedAssetsRoot is the RPM-installed root containing offline DB bootstrap assets.
	EmbeddedAssetsRoot = "/usr/share/steampipe/embedded"

	EmbeddedAssetMetadataFile = "bundle.json"
)

func EmbeddedDBAssetsDir() string {
	return filepath.Join(EmbeddedAssetsRoot, "db")
}

func EmbeddedDBBundlePath() string {
	return filepath.Join(EmbeddedDBAssetsDir(), "postgres")
}

func EmbeddedDBMetadataPath() string {
	return filepath.Join(EmbeddedDBAssetsDir(), EmbeddedAssetMetadataFile)
}

func EmbeddedFDWAssetsDir() string {
	return filepath.Join(EmbeddedAssetsRoot, "fdw")
}

func EmbeddedFDWMetadataPath() string {
	return filepath.Join(EmbeddedFDWAssetsDir(), EmbeddedAssetMetadataFile)
}
