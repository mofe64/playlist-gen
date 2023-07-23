package responses

import "time"

type APIResponse struct {
	Status    int                    `json:"status"`
	Success   bool                   `json:"success"`
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}
