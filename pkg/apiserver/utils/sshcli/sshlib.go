package sshcli

import (
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"os"
)

type SSHClient struct {
	sftpClient *sftp.Client
	session    *ssh.Session
}

func (s *SSHClient) connect() {

}

func (s *SSHClient) run_command(command string) error {
	return nil
}

func (s *SSHClient) upload_agent(localFilePath, remoteFilePath string) error {
	srcFile, err := os.Open(localFilePath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := s.sftpClient.Create(remoteFilePath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	buf := make([]byte, 1024)
	for {
		n, _ := srcFile.Read(buf)
		if n == 0 {
			break
		}
		dstFile.Write(buf)
	}

	fmt.Println("copy file to remote server finished")

	return nil
}
