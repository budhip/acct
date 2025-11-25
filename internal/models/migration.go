package models

type MigrationBucketsJournalLoadRequest struct {
	SubFolder string `json:"subFolder" validate:"required" example:"sub-folder1/sub-folder2"`
}

type MigrationLoadResponse struct {
	Kind   string `json:"kind"`
	Status string `json:"status"`
}

func NewMigrationLoadResponse(status string) *MigrationLoadResponse {
	return &MigrationLoadResponse{
		Kind:   "file",
		Status: status,
	}
}
