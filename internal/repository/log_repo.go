package repository

import (
	"llm_relay/internal/db"
	"llm_relay/internal/model"
)

func CreateLog(logEntry *model.Log) error {
	return db.DB.Create(logEntry).Error
}

func ListLogs(limit int, offset int, externalModel string) ([]model.Log, error) {
	query := db.DB.Model(&model.Log{})
	if externalModel != "" {
		query = query.Where("external_model = ?", externalModel)
	}

	var logs []model.Log
	err := query.Order("id desc").Limit(limit).Offset(offset).Find(&logs).Error
	return logs, err
}

func CountLogs(externalModel string) (int64, error) {
	query := db.DB.Model(&model.Log{})
	if externalModel != "" {
		query = query.Where("external_model = ?", externalModel)
	}
	var count int64
	err := query.Count(&count).Error
	return count, err
}

func GetLogByID(id uint) (*model.Log, error) {
	var logEntry model.Log
	err := db.DB.First(&logEntry, id).Error
	if err != nil {
		return nil, err
	}
	return &logEntry, nil
}

