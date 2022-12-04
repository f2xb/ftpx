package xftp

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type SftpConfig struct {
	Addr       string
	Username   string
	Password   string
	LastResult string

	sshClient  *ssh.Client
	sftpClient *sftp.Client
}

func (conf *SftpConfig) CreateClient(addr, username, password string) (err error) {
	conf.Addr = addr
	conf.Username = username
	conf.Password = password

	config := ssh.ClientConfig{
		User: conf.Username,
		Auth: []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		Timeout: 10 * time.Second,
	}

	if conf.sshClient, err = ssh.Dial("tcp", addr, &config); err != nil {
		return fmt.Errorf("ssh dial error : %s", err)
	}

	if conf.sftpClient, err = sftp.NewClient(conf.sshClient); err != nil {
		return fmt.Errorf("sftp new client error : %s", err)
	}
	return nil
}

func (conf *SftpConfig) makeDir(dstPath string) error {
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
		if dir, err := conf.sftpClient.Stat(baseDir); err == nil {
			if dir.IsDir() {
				continue
			}
		} else if dir == nil && err != nil {
			return fmt.Errorf("make dir [%s] error: %s", baseDir, err)
		}
		if err = conf.sftpClient.Mkdir(baseDir); err != nil {
			return fmt.Errorf("make dir [%s] error: %s", baseDir, err)
		}
	}

	return nil
}

func (conf *SftpConfig) RunShell(shell string) (string, error) {
	var (
		session *ssh.Session
		err     error
	)

	if session, err = conf.sshClient.NewSession(); err != nil {
		return "", fmt.Errorf("ssh new session error : %s", err)
	}

	if output, err := session.CombinedOutput(shell); err != nil {
		return "", fmt.Errorf("session combined output error : %s", err)
	} else {
		conf.LastResult = string(output)
	}
	return conf.LastResult, nil
}

func (conf *SftpConfig) Upload(srcPath, dstPath string) error {
	if err := conf.makeDir(dstPath); err != nil {
		return err
	}

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("open src file error: %s", err)
	}

	dstFile, err := conf.sftpClient.Create(dstPath)
	if err != nil {
		return fmt.Errorf("sftp create dst path error: %s", err)
	}

	defer func() {
		if srcFile != nil {
			_ = srcFile.Close()
		}
		if dstFile != nil {
			_ = dstFile.Close()
		}
	}()

	buf := make([]byte, 1024)
	for {
		n, err := srcFile.Read(buf)
		if err != nil {
			if err != io.EOF {
				return fmt.Errorf("read src file error : %s", err)
			} else {
				break
			}
		}
		if _, err = dstFile.Write(buf[:n]); err != nil {
			return fmt.Errorf("write dest file error: %s", err)
		}
	}

	return nil
}

func (conf *SftpConfig) Download(srcPath, dstPath string) error {
	srcFile, _ := conf.sftpClient.Open(srcPath)
	dstFile, _ := os.Create(dstPath)
	defer func() {
		_ = srcFile.Close()
		_ = dstFile.Close()
	}()

	if _, err := srcFile.WriteTo(dstFile); err != nil {
		return fmt.Errorf("write src file error : %s", err)
	}
	return nil
}
