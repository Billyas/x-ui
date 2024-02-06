package service

import (
	"gorm.io/gorm"
	"x-ui/database"
	"x-ui/database/model"
)

type SubService struct {
}

func (s *SubService) GetSubs() ([]*model.Sub, error) {
	db := database.GetDB()
	var subs []*model.Sub
	err := db.Model(model.Sub{}).Find(&subs).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return subs, nil
}
func (s *SubService) GetSubsByType(subType string) ([]*model.Sub, error) {
	db := database.GetDB()
	var subs []*model.Sub
	err := db.Model(model.Sub{}).Where("type = ?", subType).Find(&subs).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return subs, nil
}
func (s *SubService) GetSubsBySubType(subType model.SubType) ([]*model.Sub, error) {
	db := database.GetDB()
	var subs []*model.Sub
	err := db.Model(model.Sub{}).Where("type = ?", string(subType)).Find(&subs).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return subs, nil
}
func (s *SubService) AddSub(sub *model.Sub) error {
	db := database.GetDB()
	return db.Create(sub).Error
}

func (s *SubService) DelSub(id int) error {
	db := database.GetDB()
	return db.Delete(&model.Sub{}, id).Error
}

func (s *SubService) UpdateSub(sub *model.Sub) error {
	db := database.GetDB()
	return db.Save(sub).Error
}
