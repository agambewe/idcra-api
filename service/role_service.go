package service

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/kerti/idcra-api/model"
	"github.com/op/go-logging"
)

type RoleService struct {
	db  *sqlx.DB
	log *logging.Logger
}

func NewRoleService(db *sqlx.DB, log *logging.Logger) *RoleService {
	return &RoleService{db: db, log: log}
}

func (r *RoleService) FindByUserId(userId *string) ([]*model.Role, error) {
	roles := make([]*model.Role, 0)

	roleSQL := `SELECT role.*
	FROM roles role
	INNER JOIN rel_users_roles ur ON role.id = ur.role_id
	WHERE ur.user_id = ? `
	err := r.db.Select(&roles, roleSQL, userId)
	if err == sql.ErrNoRows {
		return roles, nil
	}
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *RoleService) FindRoleIdByName(roleName string) (*model.Role, error) {
	role := &model.Role{}

	roleSQL := `SELECT * FROM roles where name = ?`
	udb := r.db.Unsafe()
	row := udb.QueryRowx(roleSQL, roleName)
	err := row.StructScan(role)
	if err == sql.ErrNoRows {
		return role, nil
	}
	if err != nil {
		r.log.Errorf("Error in retrieving role : %v", err)
		return nil, err
	}

	if err == sql.ErrNoRows {
		return role, nil
	}
	if err != nil {
		return nil, err
	}

	return role, nil
}

func (r *RoleService) FindRoleId(roleId *string) (*model.Role, error) {
	role := &model.Role{}

	roleSQL := `SELECT * FROM roles where id = ?`
	udb := r.db.Unsafe()
	row := udb.QueryRowx(roleSQL, roleId)
	err := row.StructScan(role)
	if err == sql.ErrNoRows {
		return role, nil
	}
	if err != nil {
		r.log.Errorf("Error in retrieving role name: %v", err)
		return nil, err
	}

	if err == sql.ErrNoRows {
		return role, nil
	}
	if err != nil {
		return nil, err
	}

	return role, nil
}
