package config

import (
	"time"
)

//数据库配置
const (
	DBhost      = "127.0.0.1:27017"
	DBname      = "admin"
	DBUser      = "liAdmin"
	DBPwd       = "123123"
	DBtimeout   = 60 * time.Second
	DBpoollimit = 4096
	ADname      = "admin"
	ADPwd       = "123123"
	POPaccount  = "accountID"
	POPCookie   = "Cookie"
)
