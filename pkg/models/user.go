package models

import (
	"crypto/md5"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/util/uuid"
)

type User struct {
	Id       int       `xorm:"autoincr pk"`
	Uuid     string    `xorm:"unique comment('用户UUID')"`
	Name     string    `xorm:"not null unique comment('用户名')"`
	Password string    `xorm:"not null comment('密码HASH')"`
	Created  time.Time `xorm:"created"`
	Updated  time.Time `xorm:"updated"`
}

func NewUser(name, passwd string) *User {
	enPasswd := fmt.Sprintf("%x", md5.Sum([]byte(passwd)))
	return &User{
		Uuid:     string(uuid.NewUUID()),
		Name:     name,
		Password: enPasswd,
	}
}
