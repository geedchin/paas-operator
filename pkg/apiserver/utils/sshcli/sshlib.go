package sshcli

import (
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"net"
	"os"
	"time"
)

type SSHClient struct {
	Host       string
	Port       string
	Username   string
	Password   string
	Cli        *ssh.Client
	LastResult string
}

func New(ip, username, password, port string) *SSHClient {
	return &SSHClient{
		Host:     ip,
		Port:     port,
		Username: username,
		Password: password,
	}
}

func (s *SSHClient) ValidateConn() error {
	var (
		auth   []ssh.AuthMethod
		addr   string
		cliCfg *ssh.ClientConfig
		err    error
		client *ssh.Client
	)

	// auth
	auth = make([]ssh.AuthMethod, 0)
	auth = append(auth, ssh.Password(s.Password))

	cliCfg = &ssh.ClientConfig{
		User:    s.Username,
		Auth:    auth,
		Timeout: 1 * time.Minute,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	// connect
	addr = fmt.Sprintf("%s:%s", s.Host, s.Port)
	if client, err = ssh.Dial("tcp", addr, cliCfg); err != nil {
		return err
	}
	s.Cli = client
	return nil
}

func (s *SSHClient) UploadFile(localFilePath, remoteFilePath string) error {
	srcFile, err := os.Open(localFilePath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	sftpCli, err := sftp.NewClient(s.Cli)
	if err != nil {
		return err
	}
	defer sftpCli.Close()

	dstFile, err := sftpCli.Create(remoteFilePath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}
	return nil
}

func (s *SSHClient) ExecCmd(cmd string) (string, error) {
	session, err := s.Cli.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	buf, err := session.CombinedOutput(cmd)
	if err != nil {
		return "", nil
	}

	s.LastResult = string(buf)
	return s.LastResult, nil
}
