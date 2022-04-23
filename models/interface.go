package models

type MySQLInterface interface {
	CreateUser(u *User) (int64, error)
	GetUser(ID string) (*User, error)
	UpdateUser(ID string, u User) error
	DeleteUser(ID string) error
}

type RedisInterface interface {
	GetCounter() (string, error)
	SetCounter(n int) (string, error)
	Set(string, interface{}) error
	Get(string) (string, error)
}
