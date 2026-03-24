package smtp

type Repository interface {
	Create(profile *SMTPProfile) error
	Update(profile *SMTPProfile) error
	Delete(id int64) error

	GetByID(id int64) (*SMTPProfile, error)
	GetAll(userID int64) ([]SMTPProfile, error)
	ExistsByConnection(host string, port int, username string, userID int64, isGlobal bool, excludeID *int64) (bool, error)
}
