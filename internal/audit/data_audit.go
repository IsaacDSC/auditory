package audit

import (
	"fmt"
	"time"
)

type DataAudit struct {
	Metadata MetadataAudit `json:"metadata"`
	Data     any           `json:"data"`
}

type MetadataAudit struct {
	Key           string    `json:"key"` // example: "user:123"
	EventName     string    `json:"event_name"`
	RequestID     string    `json:"request_id"`
	CorrelationID string    `json:"correlation_id"`
	EventAt       time.Time `json:"event_at"`
}

func (m MetadataAudit) Validate() error {
	for _, field := range []string{m.Key, m.EventName, m.RequestID, m.CorrelationID} {
		if field == "" {
			return fmt.Errorf("field %s is required", field)
		}
	}
	return nil
}
