package capsolver

import "os"

var (
	ApiKey  = os.Getenv("CAPSOLVER_API_KEY")
	ApiHost = "https://api.capsolver.com"
)

const (
	StatusReady   = "ready"
	CreateTaskUri = "/createTask"
	GetTaskUri    = "/getTaskResult"
	BalanceUri    = "/getBalance"
	TaskTimeout   = 45
)
