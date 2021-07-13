package translater

import (
	"math/rand"
	"time"
)

type Replace struct {
	// 配置
	conf conf
}

func (re *Replace) Init() {
	rand.Seed(time.Now().UnixNano())
	intArr := rand.Perm(passwordLen)
	for key, value := range intArr {
		re.conf.EncryptPassword[key] = byte(value)
		re.conf.DecryptPassword[value] = byte(key)
	}
}

func (re *Replace) GetPW() []byte {
	return re.conf.EncryptPassword[:]
}

func (re *Replace) GenNewPW(newPW []byte) {
	for key, value := range newPW {
		re.conf.EncryptPassword[key] = byte(value)
		re.conf.DecryptPassword[value] = byte(key)
	}
}

func (re Replace) Encrypt(st []byte) []byte {
	return st
	var ret []byte
	for k, v := range st {
		ret = append(ret, re.conf.EncryptPassword[v])
		st[k] = re.conf.EncryptPassword[v]
	}
	return ret
}

func (re Replace) Decrypt(st []byte) []byte {
	return st
	var ret []byte
	for k, v := range st {
		ret = append(ret, re.conf.DecryptPassword[v])
		st[k] = re.conf.DecryptPassword[v]
	}
	return ret
}
