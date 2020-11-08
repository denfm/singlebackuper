package service

type BackupModuleResult struct {
	ArchivePath      string
	RsyncArchivePath string
	SizeMb           float64
	DurationSeconds  float64
	DurationLabel    string
	Err              error
}

type BackupModule interface {
	Backup() *BackupModuleResult
}
