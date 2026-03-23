package config

import (
	`context`
	`encoding/json`
	`errors`
	`fmt`

	`gorm.io/gorm`
)

type Repositories interface {
	GetConfigEntity(ctx context.Context, configID string) (ConfitEnt, error)
	CreateOrUpdate(ctx context.Context, configID string, configVal ConfitEnt) (ConfitEnt, error)
	Delete(ctx context.Context, configID string) error
}

type repoConfig struct {
	db *gorm.DB
}

func NewRepoConfig(dbClient *gorm.DB) Repositories {
	return &repoConfig{
		db: dbClient,
	}
}

func (d *repoConfig) GetConfigEntity(ctx context.Context, configID string) (ConfitEnt, error) {
	var config ConfitEnt

	if err := d.db.WithContext(ctx).
		Table(aegConfigTable).
		Model(ConfitEnt{}).
		Where("config_id", configID).
		First(&config).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ConfitEnt{}, fmt.Errorf("config %s not found", configID)
		}

		return ConfitEnt{}, err
	}

	return config, nil
}

func (d *repoConfig) CreateOrUpdate(ctx context.Context, configID string, configVal ConfitEnt) (ConfitEnt, error) {
	var (
		conf        ConfitEnt
		isNewConfig bool
	)

	if err := d.db.WithContext(ctx).
		Table(aegConfigTable).
		Model(ConfitEnt{}).
		Where("config_id", configID).
		First(&conf).Error; err != nil {

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return ConfitEnt{}, err
		}

		isNewConfig = true
	}

	configJSON, err := json.Marshal(configVal)
	if err != nil {
		return ConfitEnt{}, err
	}

	conf.ConfigID = configID
	conf.RawValue = string(configJSON)

	// Create config if not exist
	if isNewConfig {
		if err := d.db.WithContext(ctx).
			Table(aegConfigTable).
			Model(ConfitEnt{}).
			Create(&conf).Error; err != nil {

			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return ConfitEnt{}, err
			}
		}

		return conf, nil
	}

	// Update config if exist
	if err := d.db.WithContext(ctx).
		Table(aegConfigTable).
		Model(ConfitEnt{}).
		Where("config_id", configID).
		Save(&conf).Error; err != nil {

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return ConfitEnt{}, err
		}
	}

	return conf, nil
}

func (d *repoConfig) Delete(ctx context.Context, configID string) error {
	if err := d.db.WithContext(ctx).
		Table(aegConfigTable).
		Model(ConfitEnt{}).
		Where("config_id", configID).
		Delete(&ConfitEnt{}).Error; err != nil {

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	}

	return nil
}
