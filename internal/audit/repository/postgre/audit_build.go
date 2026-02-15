package postgre

import (
	"encoding/json"
	"identity-srv/internal/model"
)

// scanAuditLog scans a single audit log row into a model.AuditLog
func (r *implRepository) scanAuditLog(scanner interface {
	Scan(dest ...interface{}) error
}) (model.AuditLog, error) {
	var log model.AuditLog
	var metadataJSON []byte

	err := scanner.Scan(
		&log.ID,
		&log.UserID,
		&log.Action,
		&log.ResourceType,
		&log.ResourceID,
		&metadataJSON,
		&log.IPAddress,
		&log.UserAgent,
		&log.CreatedAt,
		&log.ExpiresAt,
	)
	if err != nil {
		return model.AuditLog{}, err
	}

	// Unmarshal metadata JSON
	log.Metadata = r.unmarshalMetadata(metadataJSON)

	return log, nil
}

// marshalMetadata marshals metadata map to JSON bytes
func (r *implRepository) marshalMetadata(metadata map[string]string) []byte {
	if metadata == nil {
		return []byte("{}")
	}
	data, err := json.Marshal(metadata)
	if err != nil {
		return []byte("{}")
	}
	return data
}

// unmarshalMetadata unmarshals JSON bytes to metadata map
func (r *implRepository) unmarshalMetadata(data []byte) map[string]interface{} {
	if len(data) == 0 {
		return make(map[string]interface{})
	}
	var metadata map[string]interface{}
	if err := json.Unmarshal(data, &metadata); err != nil {
		return make(map[string]interface{})
	}
	return metadata
}
