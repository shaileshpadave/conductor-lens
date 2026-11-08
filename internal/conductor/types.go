package conductor

type WorkflowStatus string

const (
	StatusFailed     WorkflowStatus = "FAILED"
	StatusCompleted  WorkflowStatus = "COMPLETED"
	StatusRunning    WorkflowStatus = "RUNNING"
	StatusTimedOut   WorkflowStatus = "TIMED_OUT"
	StatusTerminated WorkflowStatus = "TERMINATED"
)

type Workflow struct {
	WorkflowID           string         `json:"workflowId"`
	WorkflowName         string         `json:"workflowName"`
	WorkflowVersion      int            `json:"workflowVersion"`
	Status               WorkflowStatus `json:"status"`
	StartTime            int64          `json:"startTime"`
	EndTime              int64          `json:"endTime"`
	ReasonForIncompletion string        `json:"reasonForIncompletion"`
	FailedReferenceTaskNames []string   `json:"failedReferenceTaskNames"`
	Tasks                []Task         `json:"tasks"`
	Input                map[string]any `json:"input"`
	Output               map[string]any `json:"output"`
}

type Task struct {
	TaskID                string         `json:"taskId"`
	TaskType              string         `json:"taskType"`
	ReferenceTaskName     string         `json:"referenceTaskName"`
	Status                string         `json:"status"`
	ReasonForIncompletion string         `json:"reasonForIncompletion"`
	StartTime             int64          `json:"startTime"`
	EndTime               int64          `json:"endTime"`
	Retried               bool           `json:"retried"`
	RetryCount            int            `json:"retryCount"`
	CallbackAfterSeconds  int64          `json:"callbackAfterSeconds"`
	InputData             map[string]any `json:"inputData"`
	OutputData            map[string]any `json:"outputData"`
	WorkerID              string         `json:"workerId"`
}

type SearchResult struct {
	Results    []WorkflowSummary `json:"results"`
	TotalHits  int               `json:"totalHits"`
}

type WorkflowSummary struct {
	WorkflowID   string `json:"workflowId"`
	WorkflowName string `json:"workflowType"`
	Status       string `json:"status"`
	StartTime    int64  `json:"startTime"`
	EndTime      int64  `json:"endTime"`
	ReasonForIncompletion string `json:"reasonForIncompletion"`
}
