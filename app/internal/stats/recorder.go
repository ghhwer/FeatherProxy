package stats

import "FeatherProxy/app/internal/database/schema"

// Recorder is the interface for recording proxy stats asynchronously.
// Implementations must not block the caller (e.g. send to a channel or drop if full).
type Recorder interface {
	Record(stat schema.ProxyStat)
}
