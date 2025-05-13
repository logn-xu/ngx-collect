package sshclient

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/logn/ngx-collect/internal/config"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

func NewSshConnect(m config.MachineConfig) (*ssh.Client, error) {
	// var hostKey ssh.PublicKey

	// authmethod key or password
	var authMethods []ssh.AuthMethod
	if m.AuthMethod == "key" {
		// parse ~ home dir
		if strings.HasPrefix(m.KeyPath, "~") {
			m.KeyPath = strings.Replace(m.KeyPath, "~", os.Getenv("HOME"), 1)
		}
		key, err := os.ReadFile(m.KeyPath)
		if err != nil {
			log.Errorf("Read PrivateKey Failed: %v", err)
			return nil, err
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			log.Errorf("Parse PrivateKey Failed: %v", err)
			return nil, err
		}
		authMethods = []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		}

	} else if m.AuthMethod == "password" {
		authMethods = []ssh.AuthMethod{
			ssh.Password(m.Password),
		}
	}

	config := &ssh.ClientConfig{
		User:            m.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: add host key callback
		Timeout:         5 * time.Second,
	}

	conn, err := ssh.Dial("tcp", m.Host+":"+strconv.Itoa(m.Port), config)
	if err != nil {
		log.Errorf("Dial Failed: %v", err)
		return nil, err
	}

	return conn, nil
}
