package repository

import (
	"flexphish/internal/domain/group"
	"flexphish/internal/domain/target"

	"gorm.io/gorm"
)

type GroupRepository struct {
	db *gorm.DB
}

func NewGroupRepository(db *gorm.DB) group.Repository {
	return &GroupRepository{db: db}
}

func (r *GroupRepository) Create(g *group.Group) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		inputTargets := g.Targets
		g.Targets = nil

		if err := tx.Create(g).Error; err != nil {
			return err
		}

		if len(inputTargets) > 0 {
			if err := tx.Create(&inputTargets).Error; err != nil {
				return err
			}

			if err := tx.Model(g).Association("Targets").Replace(inputTargets); err != nil {
				return err
			}
		}

		return tx.Preload("Targets").First(g, g.Id).Error
	})
}

func (r *GroupRepository) Update(g *group.Group) error {
	return r.db.Save(g).Error
}

func (r *GroupRepository) Delete(id int64) error {
	return r.db.Delete(&group.Group{}, id).Error
}

func (r *GroupRepository) GetByID(id int64) (*group.Group, error) {
	var g group.Group
	err := r.db.Preload("Targets").First(&g, id).Error
	return &g, err
}

func (r *GroupRepository) GetAll(userID int64) ([]group.Group, error) {
	var groups []group.Group

	err := r.db.
		Where("is_global = ?", true).
		Or("user_id = ?", userID).
		Preload("Targets").
		Find(&groups).Error

	return groups, err
}

func (r *GroupRepository) ExistsByName(name string, userID int64, isGlobal bool, excludeID *int64) (bool, error) {
	var count int64

	query := r.db.Model(&group.Group{}).Where("LOWER(name) = LOWER(?)", name)
	if excludeID != nil {
		query = query.Where("id <> ?", *excludeID)
	}

	if isGlobal {
		query = query.Where("is_global = ?", true)
	} else {
		query = query.Where("is_global = ? OR user_id = ?", true, userID)
	}

	if err := query.Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *GroupRepository) ListTargets(groupID int64) ([]target.Target, error) {
	var targets []target.Target

	err := r.db.
		Table("targets").
		Joins("JOIN group_targets ON group_targets.target_id = targets.id").
		Where("group_targets.group_id = ?", groupID).
		Order("targets.id DESC").
		Find(&targets).Error
	if err != nil {
		return nil, err
	}

	return targets, nil
}

func (r *GroupRepository) GetTargetByID(groupID int64, targetID int64) (*target.Target, error) {
	var t target.Target

	err := r.db.
		Table("targets").
		Joins("JOIN group_targets ON group_targets.target_id = targets.id").
		Where("group_targets.group_id = ? AND targets.id = ?", groupID, targetID).
		First(&t).Error
	if err != nil {
		return nil, err
	}

	return &t, nil
}

func (r *GroupRepository) TargetEmailExistsInGroup(groupID int64, email string, excludeTargetID *int64) (bool, error) {
	var count int64

	query := r.db.
		Table("targets").
		Joins("JOIN group_targets ON group_targets.target_id = targets.id").
		Where("group_targets.group_id = ? AND LOWER(targets.email) = LOWER(?)", groupID, email)

	if excludeTargetID != nil {
		query = query.Where("targets.id <> ?", *excludeTargetID)
	}

	if err := query.Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *GroupRepository) CreateTarget(groupID int64, targetData *target.Target) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(targetData).Error; err != nil {
			return err
		}

		g := group.Group{Id: groupID}
		if err := tx.Model(&g).Association("Targets").Append(targetData); err != nil {
			return err
		}

		return nil
	})
}

func (r *GroupRepository) UpdateTarget(groupID int64, targetData *target.Target) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Table("group_targets").
			Where("group_id = ? AND target_id = ?", groupID, targetData.Id).
			Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return gorm.ErrRecordNotFound
		}

		if err := tx.Model(&target.Target{}).
			Where("id = ?", targetData.Id).
			Updates(map[string]interface{}{
				"first_name": targetData.FirstName,
				"last_name":  targetData.LastName,
				"email":      targetData.Email,
				"position":   targetData.Position,
			}).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *GroupRepository) DeleteTarget(groupID int64, targetID int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Table("group_targets").
			Where("group_id = ? AND target_id = ?", groupID, targetID).
			Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return gorm.ErrRecordNotFound
		}

		g := group.Group{Id: groupID}
		t := target.Target{Id: targetID}

		if err := tx.Model(&g).Association("Targets").Delete(&t); err != nil {
			return err
		}

		var remaining int64
		if err := tx.Table("group_targets").
			Where("target_id = ?", targetID).
			Count(&remaining).Error; err != nil {
			return err
		}
		if remaining == 0 {
			if err := tx.Delete(&target.Target{}, targetID).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
