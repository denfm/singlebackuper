package service

type BackupModuleResult struct {
	ArchivePath       string
	RemoteArchivePath string
	SizeMb            float64
	SizeMbLabel       string
	DurationSeconds   float64
	DurationLabel     string
	Err               error
}

type BackupModule interface {
	Backup() *BackupModuleResult
}
