package translater

type replace struct {
	// 配置
	conf conf
}

type conf struct {
	// 加密密码
	EncryptPassword [256]byte
	DecryptPassword [256]byte
}

// 初始化配置文件
func (re *replace) Init(confFile string) {
	// read conf file
	re.conf = conf{}
}

func (re replace) TransLater(st []byte) []byte {
	var ret []byte
	for _, v := range st {
		ret = append(ret, re.conf.EncryptPassword[v])
	}
	return ret
}