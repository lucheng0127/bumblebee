package authorizer

import (
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	xormadapter "github.com/casbin/xorm-adapter/v3"
	_ "github.com/mattn/go-sqlite3"
)

// Policy obj: /api/<group>/<version>/* act: GET, POST, PUT, PATCH, DELETE
const (
	AdminSubPrefix = "global-"

	AdminPolicyObj = "/api/*"

	PolicyActGet    = "GET"
	PolicyActPost   = "POST"
	PolicyActPut    = "PUT"
	PolicyActPatch  = "PATCH"
	PolicyActDELETE = "DELETE"
)

type DBAuthorizer struct {
	database string
	enforcer *casbin.Enforcer
}

func NewDBAuthorizer(database string) (*DBAuthorizer, error) {
	a := &DBAuthorizer{
		database: database,
	}

	adp, err := xormadapter.NewAdapter("sqlite3", database)
	if err != nil {
		return nil, err
	}

	m, err := model.NewModelFromString(`
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && keyMatch5(r.obj, p.obj) && r.act == p.act
`)
	if err != nil {
		return nil, err
	}

	e, err := casbin.NewEnforcer(m, adp)
	if err != nil {
		return nil, err
	}

	a.enforcer = e

	return a, nil
}

type RolePolicy struct {
	RoleName string
	Url      string
	Method   string
}

func (a *DBAuthorizer) GetRoles() ([]string, error) {
	return a.enforcer.GetAllRoles()
}

func (a *DBAuthorizer) CreateRolePolicy(p RolePolicy) error {
	_, err := a.enforcer.AddPolicy(p.RoleName, p.Url, p.Method)
	if err != nil {
		return err
	}

	return a.enforcer.SavePolicy()
}

func (a *DBAuthorizer) DeleteRolePolicy(p RolePolicy) error {
	_, err := a.enforcer.RemovePolicy(p.RoleName, p.Url, p.Method)
	if err != nil {
		return err
	}

	return a.enforcer.SavePolicy()
}

func (a *DBAuthorizer) AddUserRole(user, role string) error {
	_, err := a.enforcer.AddGroupingPolicy(user, role)
	if err != nil {
		return err
	}

	return a.enforcer.SavePolicy()
}

func (a *DBAuthorizer) DeleteUserRole(user, role string) error {
	_, err := a.enforcer.RemoveGroupingPolicy(user, role)
	if err != nil {
		return err
	}

	return a.enforcer.SavePolicy()
}

func (a *DBAuthorizer) Authenticate(user, url, method string) (bool, error) {
	if err := a.enforcer.LoadPolicy(); err != nil {
		return false, err
	}

	return a.enforcer.Enforce(user, url, method)
}
