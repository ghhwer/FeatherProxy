package impl

import (
	"time"

	"FeatherProxy/app/internal/database/objects"
	"FeatherProxy/app/internal/database/schema"

	"github.com/google/uuid"
)

func (r *repository) CreateProxyStats(stats []schema.ProxyStat) error {
	if len(stats) == 0 {
		return nil
	}
	objs := make([]objects.ProxyStat, len(stats))
	for i := range stats {
		s := &stats[i]
		if s.ID == uuid.Nil {
			s.ID = uuid.New()
		}
		objs[i] = objects.SchemaToProxyStat(*s)
	}
	return r.db.CreateInBatches(objs, 100).Error
}

func (r *repository) ListProxyStats(limit, offset int, since *time.Time) ([]schema.ProxyStat, int64, error) {
	var total int64
	q := r.db.Model(&objects.ProxyStat{})
	if since != nil {
		q = q.Where("timestamp >= ?", *since)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []objects.ProxyStat
	q = r.db.Order("timestamp DESC").Limit(limit).Offset(offset)
	if since != nil {
		q = q.Where("timestamp >= ?", *since)
	}
	if err := q.Find(&list).Error; err != nil {
		return nil, 0, err
	}
	out := make([]schema.ProxyStat, len(list))
	for i := range list {
		out[i] = objects.ProxyStatToSchema(&list[i])
	}
	return out, total, nil
}

func (r *repository) DeleteProxyStatsOlderThan(until time.Time) error {
	return r.db.Where("timestamp < ?", until).Delete(&objects.ProxyStat{}).Error
}

func (r *repository) ClearAllProxyStats() error {
	return r.db.Where("1 = 1").Delete(&objects.ProxyStat{}).Error
}

func (r *repository) StatsSummary() (schema.StatsSummary, error) {
	var out schema.StatsSummary
	now := time.Now()
	last24h := now.Add(-24 * time.Hour)
	last1min := now.Add(-1 * time.Minute)

	if err := r.db.Model(&objects.ProxyStat{}).Count(&out.Total).Error; err != nil {
		return out, err
	}
	if err := r.db.Model(&objects.ProxyStat{}).Where("timestamp >= ?", last24h).Count(&out.Last24h).Error; err != nil {
		return out, err
	}
	if err := r.db.Model(&objects.ProxyStat{}).Where("timestamp >= ?", last24h).Where("status_code >= ? AND status_code < ?", 200, 300).Count(&out.Status2xx).Error; err != nil {
		return out, err
	}
	if err := r.db.Model(&objects.ProxyStat{}).Where("timestamp >= ?", last24h).Where("status_code >= ? AND status_code < ?", 400, 500).Count(&out.Status4xx).Error; err != nil {
		return out, err
	}
	if err := r.db.Model(&objects.ProxyStat{}).Where("timestamp >= ?", last24h).Where("status_code >= ? AND status_code < ?", 500, 600).Count(&out.Status5xx).Error; err != nil {
		return out, err
	}
	if err := r.db.Model(&objects.ProxyStat{}).Where("timestamp >= ?", last1min).Count(&out.TpsLastMinute).Error; err != nil {
		return out, err
	}
	return out, nil
}

func (r *repository) StatsByRoute(since *time.Time, limit int) ([]schema.RouteCount, error) {
	q := r.db.Model(&objects.ProxyStat{}).Select("route_uuid, method, path as source_path, count(*) as count").Group("route_uuid, method, path").Order("count DESC")
	if since != nil {
		q = q.Where("timestamp >= ?", *since)
	}
	if limit > 0 {
		q = q.Limit(limit)
	}
	var rows []struct {
		RouteUUID  uuid.UUID
		Method     string
		SourcePath string
		Count      int64
	}
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]schema.RouteCount, len(rows))
	for i := range rows {
		out[i] = schema.RouteCount{RouteUUID: rows[i].RouteUUID, Method: rows[i].Method, SourcePath: rows[i].SourcePath, Count: rows[i].Count}
	}
	return out, nil
}

func (r *repository) StatsByCaller(since *time.Time, limit int) ([]schema.CallerCount, error) {
	q := r.db.Model(&objects.ProxyStat{}).Select("client_ip, count(*) as count").Group("client_ip").Order("count DESC")
	if since != nil {
		q = q.Where("timestamp >= ?", *since)
	}
	if limit > 0 {
		q = q.Limit(limit)
	}
	var rows []struct {
		ClientIP string
		Count    int64
	}
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]schema.CallerCount, len(rows))
	for i := range rows {
		out[i] = schema.CallerCount{ClientIP: rows[i].ClientIP, Count: rows[i].Count}
	}
	return out, nil
}

func (r *repository) StatsBySourceServer(since *time.Time) ([]schema.ServerCount, error) {
	q := r.db.Model(&objects.ProxyStat{}).Select("source_server_uuid as server_uuid, count(*) as count").Group("source_server_uuid").Order("count DESC")
	if since != nil {
		q = q.Where("timestamp >= ?", *since)
	}
	var rows []struct {
		ServerUUID uuid.UUID
		Count      int64
	}
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]schema.ServerCount, len(rows))
	for i := range rows {
		out[i] = schema.ServerCount{ServerUUID: rows[i].ServerUUID, Count: rows[i].Count}
	}
	return out, nil
}

func (r *repository) StatsByTargetServer(since *time.Time) ([]schema.ServerCount, error) {
	q := r.db.Model(&objects.ProxyStat{}).Select("target_server_uuid as server_uuid, count(*) as count").Group("target_server_uuid").Order("count DESC")
	if since != nil {
		q = q.Where("timestamp >= ?", *since)
	}
	var rows []struct {
		ServerUUID uuid.UUID
		Count      int64
	}
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]schema.ServerCount, len(rows))
	for i := range rows {
		out[i] = schema.ServerCount{ServerUUID: rows[i].ServerUUID, Count: rows[i].Count}
	}
	return out, nil
}

func (r *repository) StatsTPS(since time.Time, bucketDuration time.Duration) ([]schema.BucketCount, error) {
	// Use raw SQL for time bucketing: PostgreSQL date_trunc, SQLite strftime
	dialect := r.db.Dialector.Name()
	var rows []struct {
		At    time.Time
		Count int64
	}
	// Query: group by bucket, order by bucket
	switch dialect {
	case "postgres", "postgresql":
		err := r.db.Raw(`
			SELECT date_trunc('minute', timestamp) AS at, COUNT(*) AS count
			FROM proxy_stats WHERE timestamp >= ? AND timestamp <= ?
			GROUP BY 1 ORDER BY 1`,
			since, time.Now()).Scan(&rows).Error
		if err != nil {
			return nil, err
		}
	case "sqlite":
		var sqliteRows []struct {
			AtStr string `gorm:"column:at_str"`
			Count int64  `gorm:"column:count"`
		}
		err := r.db.Raw(`
			SELECT strftime('%Y-%m-%d %H:%M:00', timestamp) AS at_str, COUNT(*) AS count
			FROM proxy_stats WHERE timestamp >= ? AND timestamp <= ?
			GROUP BY at_str ORDER BY at_str`, since, time.Now()).Scan(&sqliteRows).Error
		if err != nil {
			return nil, err
		}
		rows = make([]struct {
			At    time.Time
			Count int64
		}, len(sqliteRows))
		for i := range sqliteRows {
			t, _ := time.ParseInLocation("2006-01-02 15:04:05", sqliteRows[i].AtStr, time.UTC)
			rows[i] = struct {
				At    time.Time
				Count int64
			}{At: t, Count: sqliteRows[i].Count}
		}
	default:
		// Fallback: same as postgres
		err := r.db.Raw(`
			SELECT date_trunc('minute', timestamp) AS at, COUNT(*) AS count
			FROM proxy_stats WHERE timestamp >= ? AND timestamp <= ?
			GROUP BY 1 ORDER BY 1`,
			since, time.Now()).Scan(&rows).Error
		if err != nil {
			return nil, err
		}
	}
	out := make([]schema.BucketCount, len(rows))
	for i := range rows {
		out[i] = schema.BucketCount{At: rows[i].At, Count: rows[i].Count}
	}
	return out, nil
}
