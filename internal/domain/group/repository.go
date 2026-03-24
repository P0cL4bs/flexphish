package group

import "flexphish/internal/domain/target"

type Repository interface {
	Create(group *Group) error
	Update(group *Group) error
	Delete(id int64) error

	GetByID(id int64) (*Group, error)
	GetAll(userID int64) ([]Group, error)
	ExistsByName(name string, userID int64, isGlobal bool, excludeID *int64) (bool, error)

	ListTargets(groupID int64) ([]target.Target, error)
	GetTargetByID(groupID int64, targetID int64) (*target.Target, error)
	TargetEmailExistsInGroup(groupID int64, email string, excludeTargetID *int64) (bool, error)
	CreateTarget(groupID int64, targetData *target.Target) error
	UpdateTarget(groupID int64, targetData *target.Target) error
	DeleteTarget(groupID int64, targetID int64) error
}
