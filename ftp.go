package xftp

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jlaffaye/ftp"
)

type FtpConfig struct {
	Addr     string
	Username string
	Password string

	conn *ftp.ServerConn
}

func (conf *FtpConfig) quit() {
	_ = conf.conn.Quit()
}

func (conf *FtpConfig) CreateClient(addr, username, password string) (err error) {
	conf.Addr = addr
	conf.Username = username
	conf.Password = password

	conf.conn, err = ftp.Dial(addr, ftp.DialWithTimeout(10*time.Second))
	if err != nil {
		return fmt.Errorf("ftp dial error : %s", err)
	}

	if err = conf.conn.Login(username, password); err != nil {
		return fmt.Errorf("ftp login error : %s", err)
	}

	return nil
}

func (conf *FtpConfig) makeDir(dstPath string) error {
	dirs, err := checkPath(dstPath)
	if err != nil {
		return err
	}
	if len(dirs) == 0 {
		return nil
	}

	for i, _ := range dirs {
		baseDir := filepath.Join(dirs[:i+1]...)
		baseDir = strings.ReplaceAll(baseDir, `\\`, `/`)
		baseDir = strings.ReplaceAll(baseDir, `\`, `/`)
		if err := conf.conn.MakeDir(baseDir); err != nil {
			if !strings.Contains(err.Error(), "already exists") {
				return fmt.Errorf("make dir [%s] error: %s", baseDir, err)
			}
		}
	}

	return nil
}

func (conf *FtpConfig) Upload(srcPath, dstPath string) error {
	defer conf.quit()

	if err := conf.makeDir(dstPath); err != nil {
		return err
	}

	bts, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("open src file error: %s", err)
	}

	if err = conf.conn.Stor(dstPath, bts); err != nil {
		return fmt.Errorf("upload file error: %s", err)
	}

	return nil
}

func (conf *FtpConfig) Download(srcPath, dstPath string) error {
	defer conf.quit()

	resp, err := conf.conn.Retr(srcPath)
	if err != nil {
		return fmt.Errorf("download file error: %s", err)
	}

	bts, err := ioutil.ReadAll(resp)
	if err != nil {
		return fmt.Errorf("read file error: %s", err)
	}

	return ioutil.WriteFile(dstPath, bts, os.ModePerm)
}
