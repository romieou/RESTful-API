package mysql

import (
	"database/sql"
	"log"
	"rest/models"
	"rest/myerrors"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type MySQL struct {
	db *sql.DB
}

// NewMySQL return new instance of MySQL
func NewMySQL() (models.MySQLInterface, error) {
	db, err := sql.Open("mysql", "tester:secret@tcp(db:3306)/db")
	if err != nil {
		return nil, err
	}
	log.Println("INFO|Success in opening DB")
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(time.Minute * 5)
	// db.SetConnMaxIdleTime(time.Minute * 2)
	start := time.Now()
	for db.Ping() != nil {
		if time.Now().After(start.Add(time.Minute * 20)) {
			log.Println("ERROR|Failed to connect after 20 minutes")
			return nil, db.Ping()
		}
	}
	log.Println("INFO|DB Pong", db.Ping() == nil)
	return &MySQL{db: db}, nil
}

// CreateUser creates adds record of new user to database and returns their ID
func (m *MySQL) CreateUser(u *models.User) (int64, error) {
	insForm, err := m.db.Prepare("INSERT INTO users (firstname, lastname) VALUES(?, ?)")
	if err != nil {
		return 0, err
	}
	res, err := insForm.Exec(u.FirstName, u.LastName)
	if err != nil {
		return 0, err
	}
	ID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return ID, nil
}

// GetUser retrieves info on user by ID
func (m *MySQL) GetUser(ID string) (*models.User, error) {
	user := new(models.User)
	err := m.db.QueryRow("SELECT * FROM users WHERE id=?", ID).Scan(&user.ID, &user.FirstName, &user.LastName)
	if err == sql.ErrNoRows {
		return user, myerrors.ErrUserNotFound
	}
	return user, err
}

// UpdateUser updates user by ID to provided input
func (m *MySQL) UpdateUser(ID string, u models.User) error {
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	if firstname := u.FirstName; firstname != "" {
		res, err := tx.Exec("UPDATE users SET firstname = ? where id = ?", firstname, ID)
		if err != nil {
			tx.Rollback()
			return err
		}
		rows, err := res.RowsAffected()
		if err != nil {
			tx.Rollback()
			return err
		}
		if rows == 0 {
			tx.Rollback()
			return myerrors.ErrUserNotFound
		}
	}

	if lastname := u.LastName; lastname != "" {
		res, err := tx.Exec("UPDATE users SET lastname = ? where id = ?", lastname, ID)
		if err != nil {
			tx.Rollback()
			return err
		}
		rows, err := res.RowsAffected()
		if err != nil {
			tx.Rollback()
			return err
		}
		if rows == 0 {
			tx.Rollback()
			return myerrors.ErrUserNotFound
		}
	}
	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

// DeleteUser deletes user by ID
// Initial user gets ID of 1
// ID does not get reset with delete
func (m *MySQL) DeleteUser(ID string) error {
	res, err := m.db.Exec("DELETE FROM users WHERE id=?", ID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return myerrors.ErrUserNotFound
	}
	return nil
}
