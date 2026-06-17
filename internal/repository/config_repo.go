package repository

import (
	"llm_relay/internal/db"
	"llm_relay/internal/model"
)

func GetAllConfigs() ([]model.Config, error) {
	var configs []model.Config
	err := db.DB.Find(&configs).Error
	return configs, err
}

func GetConfigByID(id uint) (*model.Config, error) {
	var config model.Config
	err := db.DB.First(&config, id).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func GetConfigByExternalModel(externalModel string) (*model.Config, error) {
	var config model.Config
	err := db.DB.Where("external_model = ?", externalModel).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func CreateConfig(config *model.Config) error {
	return db.DB.Create(config).Error
}

func UpdateConfig(config *model.Config) error {
	return db.DB.Save(config).Error
}

func DeleteConfig(id uint) error {
	return db.DB.Delete(&model.Config{}, id).Error
}
