package domain

import "time"

type User struct {
	Id       int64
	Email    string
	Password string
	NickName string
	Birthday string
	Profile  string

	// UTC 0 的时区
	Ctime time.Time

	//Addr Address
}
