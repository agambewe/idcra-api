package service

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/kerti/idcra-api/context"
	"github.com/kerti/idcra-api/model"
	"github.com/op/go-logging"
	uuid "github.com/satori/go.uuid"
)

const (
	defaultListFetchSize = 10
)

type UserService struct {
	db              *sqlx.DB
	roleService     *RoleService
	studentsService *StudentService
	log             *logging.Logger
}

func NewUserService(db *sqlx.DB, roleService *RoleService, studentsService *StudentService, log *logging.Logger) *UserService {
	return &UserService{db: db, roleService: roleService, studentsService: studentsService, log: log}
}

func (u *UserService) FindByEmail(email string) (*model.User, error) {
	user := &model.User{}

	userSQL := `SELECT * FROM users WHERE email = ?`
	udb := u.db.Unsafe()
	row := udb.QueryRowx(userSQL, email)
	err := row.StructScan(user)
	if err == sql.ErrNoRows {
		return user, nil
	}
	if err != nil {
		u.log.Errorf("Error in retrieving user : %v", err)
		return nil, err
	}

	roles, err := u.roleService.FindByUserId(&user.ID)
	if err != nil {
		u.log.Errorf("Error in retrieving roles : %v", err)
		return nil, err
	}
	user.Roles = roles

	students, err := u.studentsService.FindByUserId(&user.ID)
	if err != nil {
		u.log.Errorf("Error in retrieving roles : %v", err)
		return nil, err
	}
	user.Students = students

	return user, nil
}

func (u *UserService) FindUserById(userId string) (*model.User, error) {
	user := &model.User{}

	userSQL := `SELECT * FROM users WHERE id = ?`
	udb := u.db.Unsafe()
	row := udb.QueryRowx(userSQL, userId)
	err := row.StructScan(user)
	if err == sql.ErrNoRows {
		return user, nil
	}
	if err != nil {
		u.log.Errorf("Error in retrieving user : %v", err)
		return nil, err
	}

	roles, err := u.roleService.FindByUserId(&user.ID)
	if err != nil {
		u.log.Errorf("Error in retrieving roles : %v", err)
		return nil, err
	}
	user.Roles = roles

	students, err := u.studentsService.FindByUserId(&user.ID)
	if err != nil {
		u.log.Errorf("Error in retrieving roles : %v", err)
		return nil, err
	}
	user.Students = students

	return user, nil
}

func (u *UserService) CreateUser(user *model.User) (*model.User, error) {
	userID := uuid.NewV4()
	user.ID = userID.String()
	userSQL := `INSERT INTO users (id, email, password, ip_address) VALUES (:id, :email, :password, :ip_address)`

	if err := user.HashedPassword(); err != nil {
		return nil, err
	}

	if _, err := u.db.NamedExec(userSQL, user); err != nil {
		u.log.Errorf("Error in creating user : %v", err)
		return nil, err
	}

	userResult, err := u.FindByEmail(user.Email)
	if err != nil {
		u.log.Errorf("Error in retrieving user : %v", err)
		return nil, err
	}

	roleResult, err := u.roleService.FindRoleIdByName("PARENT")
	if err != nil {
		u.log.Errorf("Error in retrieving role : %v", err)
		return nil, err
	}

	roleSQL := `INSERT INTO rel_users_roles (user_id, role_id) VALUES (?,?)`

	if _, err := u.db.Exec(roleSQL, userResult.ID, roleResult.ID); err != nil {
		u.log.Errorf("Error in creating role user relation : %v", err)
		return nil, err
	}

	return userResult, nil
}

func (u *UserService) CreateUserStudentRelation(relations *model.UsersStudentsRelations) (*model.User, error) {
	userID := uuid.NewV4()
	newID := userID.String()

	userSQL := `INSERT INTO rel_users_students (id, user_id, student_id) VALUES (?, ?, ?)`

	if _, err := u.db.Exec(userSQL, newID, relations.UserId, relations.StudentId); err != nil {
		u.log.Errorf("Error in adding student to user : %v", err)
		return nil, err
	}

	userResult, err := u.FindUserById(relations.UserId)
	if err != nil {
		u.log.Errorf("Error in retrieving user : %v", err)
		return nil, err
	}

	return userResult, nil
}

func (u *UserService) DeleteStudentFromParent(relations *model.UsersStudentsRelations) (*model.User, error) {

	userSQL := `DELETE FROM rel_users_students WHERE user_id = ? AND student_id = ?`

	if _, err := u.db.Exec(userSQL, relations.UserId, relations.StudentId); err != nil {
		u.log.Errorf("Error in deleting student from user : %v", err)
		return nil, err
	}

	userResult, err := u.FindUserById(relations.UserId)
	if err != nil {
		u.log.Errorf("Error in retrieving user : %v", err)
		return nil, err
	}

	return userResult, nil
}

func (u *UserService) FindUserRole(userId *string) (*string, error) {
	user := &model.UsersRolesRelations{}

	userSQL := `SELECT * FROM rel_users_roles WHERE user_id = ?`
	udb := u.db.Unsafe()
	row := udb.QueryRowx(userSQL, userId)
	err := row.StructScan(user)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		u.log.Errorf("Error in retrieving user role : %v", err)
		return nil, err
	}

	roleResult, err := u.roleService.FindRoleId(&user.RoleId)
	if err != nil {
		u.log.Errorf("Error in retrieving role : %v", err)
		return nil, err
	}

	return &roleResult.Name, nil
}

func (u *UserService) List(first *int32, after *string) ([]*model.User, error) {
	users := make([]*model.User, 0)
	var fetchSize int32
	if first == nil {
		fetchSize = defaultListFetchSize
	} else {
		fetchSize = *first
	}

	if after != nil {
		userSQL := `SELECT * FROM users WHERE created_at < (SELECT created_at FROM users WHERE id = ?) ORDER BY created_at DESC LIMIT ?;`
		decodedIndex, _ := DecodeCursor(after)
		err := u.db.Select(&users, userSQL, decodedIndex, fetchSize)
		if err != nil {
			return nil, err
		}
		return users, nil
	}
	userSQL := `SELECT * FROM users ORDER BY created_at DESC LIMIT ?;`
	err := u.db.Select(&users, userSQL, fetchSize)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (u *UserService) Count() (int, error) {
	var count int
	userSQL := `SELECT count(*) FROM users`
	err := u.db.Get(&count, userSQL)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (u *UserService) ComparePassword(userCredentials *model.UserCredentials) (*model.User, error) {
	user, err := u.FindByEmail(userCredentials.Email)
	if err != nil {
		return nil, errors.New(context.UnauthorizedAccess)
	}
	if result := user.ComparePassword(userCredentials.Password); !result {
		return nil, errors.New(context.UnauthorizedAccess)
	}
	return user, nil
}
