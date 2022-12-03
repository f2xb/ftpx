package xftp

import (
	"testing"
)

func TestFtp(t *testing.T) {
	cliConf := new(FtpConfig)
	_ = cliConf.CreateClient("ftp.example.org:21", "root", "xxxxxx")
	// 本地文件上传到服务器
	_ = cliConf.Upload(`/app/haha1.go`, `/root/haha.go`)
	// 从服务器中下载文件
	_ = cliConf.Download(`/root/1.py`, `/app/go1.py`)
}
