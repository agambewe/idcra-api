package model

type UsersRolesRelations struct {
	UserId string `db:"user_id"`
	RoleId string `db:"role_id"`
}
