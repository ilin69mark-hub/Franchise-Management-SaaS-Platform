package jobs

import (
	"log"
	"time"

	"gorm.io/gorm"
)

type LogRotationJob struct {
	db *gorm.DB
}

func NewLogRotationJob(db *gorm.DB) *LogRotationJob {
	return &LogRotationJob{db: db}
}

// Run удаляет user_logs старше 90 дней — защита от безграничного роста + PII минимизация
func (j *LogRotationJob) Run() {
	log.Println("Running Log Rotation Job (cleanup >90d)...")
	cutoff := time.Now().AddDate(0, 0, -90)
	res := j.db.Exec("DELETE FROM user_logs WHERE created_at < ?", cutoff)
	if res.Error != nil {
		log.Printf("LogRotation error: %v", res.Error)
		return
	}
	if res.RowsAffected > 0 {
		log.Printf("LogRotation: deleted %d old user_logs", res.RowsAffected)
	}
}
