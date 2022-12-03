package xftp

import (
	"fmt"
	"io"
	"net"
	"os"
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

func (cliConf *SftpConfig) CreateClient(addr, username, password string) error {
	var (
		sshClient  *ssh.Client
		sftpClient *sftp.Client
		err        error
	)
	cliConf.Addr = addr
	cliConf.Username = username
	cliConf.Password = password

	config := ssh.ClientConfig{
		User: cliConf.Username,
		Auth: []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		Timeout: 10 * time.Second,
	}

	if sshClient, err = ssh.Dial("tcp", addr, &config); err != nil {
		return fmt.Errorf("ssh dial error : %s", err)
	}
	cliConf.sshClient = sshClient

	if sftpClient, err = sftp.NewClient(sshClient); err != nil {
		return fmt.Errorf("sftp new client error : %s", err)
	}
	cliConf.sftpClient = sftpClient
	return nil
}

func (cliConf *SftpConfig) RunShell(shell string) (string, error) {
	var (
		session *ssh.Session
		err     error
	)

	if session, err = cliConf.sshClient.NewSession(); err != nil {
		return "", fmt.Errorf("ssh new session error : %s", err)
	}

	if output, err := session.CombinedOutput(shell); err != nil {
		return "", fmt.Errorf("session combined output error : %s", err)
	} else {
		cliConf.LastResult = string(output)
	}
	return cliConf.LastResult, nil
}

func (cliConf *SftpConfig) Upload(srcPath, dstPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("open src file error: %s", err)
	}
	dstFile, err := cliConf.sftpClient.Create(dstPath)
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
		_, _ = dstFile.Write(buf[:n])
	}

	_, err = cliConf.RunShell(fmt.Sprintf("ls %s", dstPath))
	return err
}

func (cliConf *SftpConfig) Download(srcPath, dstPath string) error {
	srcFile, _ := cliConf.sftpClient.Open(srcPath)
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
