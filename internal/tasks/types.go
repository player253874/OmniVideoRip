package tasks

type Status string

const (
	StatusQueued               Status = "queued"
	StatusCheckingDependencies Status = "checking_dependencies"
	StatusWaitingForSidecars   Status = "waiting_for_sidecars"
	StatusRunning              Status = "running"
	StatusPaused               Status = "paused"
	StatusCompleted            Status = "completed"
	StatusFailed               Status = "failed"
)

type CreateTaskRequest struct {
	URL           string `json:"url"`
	OutputDir     string `json:"outputDir,omitempty"`
	Source        string `json:"source,omitempty"`
	ClientVersion string `json:"clientVersion,omitempty"`
	FormatID      string `json:"formatId,omitempty"`
}

type Task struct {
	ID            string `json:"id"`
	URL           string `json:"url"`
	Platform      string `json:"platform"`
	Status        Status `json:"status"`
	Message       string `json:"message"`
	Progress      int    `json:"progress"`
	OutputDir     string `json:"outputDir,omitempty"`
	OutputPath    string `json:"outputPath,omitempty"`
	ErrorCode     string `json:"errorCode,omitempty"`
	Source        string `json:"source,omitempty"`
	ClientVersion string `json:"clientVersion,omitempty"`
	FormatID      string `json:"formatId,omitempty"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}
