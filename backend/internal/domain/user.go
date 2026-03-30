package domain

// User represents the users table per spec Section 11.1.
type User struct {
	BaseModel
	Username     string `gorm:"type:varchar(64);not null;uniqueIndex" json:"username"`
	PasswordHash string `gorm:"type:varchar(255);not null" json:"-"`
	Nickname     string `gorm:"type:varchar(64);not null" json:"nickname"`
	Role         string `gorm:"type:varchar(32);not null;default:'editor'" json:"role"`
	Status       int16  `gorm:"type:smallint;not null;default:1" json:"status"`
}

// TableName overrides the default table name.
func (User) TableName() string {
	return "users"
}

// User role constants.
const (
	RoleAdmin  = "admin"
	RoleEditor = "editor"
)
