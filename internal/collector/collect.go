package collector

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/logn/ngx-collect/internal/config"
	"github.com/logn/ngx-collect/internal/sshclient"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type Collecter interface {
	Connect() error
	Fetch() error
	Close() error
}

type CollecterImpl struct {
	config        config.MachineConfig
	sshClient     *ssh.Client
	localBasePath string
}

func NewCollecter(config config.MachineConfig) Collecter {
	var baseDir = fmt.Sprintf("%s/%s/%s", config.LocalDestination, config.Alias, config.Host)

	return &CollecterImpl{
		config:        config,
		localBasePath: baseDir,
	}
}

// connect to server
func (c *CollecterImpl) Connect() error {
	sshClient, err := sshclient.NewSshConnect(c.config)
	if err != nil {
		return err
	}

	c.sshClient = sshClient
	return nil
}

// fetch nginx config by executing nginx -T command
func (c *CollecterImpl) Fetch() error {
	// create local base path
	if err := c.CreateBaseDestDir(); err != nil {
		log.Errorf("create base dir failed: %v", err)
		return err
	}

	// execute nginx -T -c <config_file> to get running configuration
	cmd := fmt.Sprintf("%s -T -c %s", c.config.NginxExecBin, c.config.NginxMainConfig)
	log.Infof("Executing command: %s", cmd)

	// create SSH session
	session, err := c.sshClient.NewSession()
	if err != nil {
		log.Errorf("create SSH session failed: %v", err)
		return err
	}
	defer session.Close()

	// execute command and get output
	output, err := session.CombinedOutput(cmd)
	if err != nil {
		log.Errorf("execute nginx command failed: %v", err)
		return err
	}

	// save output to local nginx.conf file
	localConfigPath := filepath.Join(c.localBasePath, "nginx.conf")
	if err := os.WriteFile(localConfigPath, output, 0644); err != nil {
		log.Errorf("write nginx config file failed: %v", err)
		return err
	}

	log.Infof("Nginx configuration saved to: %s", localConfigPath)
	return nil
}

func (c *CollecterImpl) Close() error {
	if c.sshClient != nil {
		return c.sshClient.Close()
	}
	return nil
}

// create base dir
// format: local_destination/alias/host
func (c *CollecterImpl) CreateBaseDestDir() error {
	err := os.MkdirAll(fmt.Sprintf("%s/%s/%s", c.config.LocalDestination, c.config.Alias, c.config.Host), 0755)
	if err != nil {
		log.Errorf("create local dir failed: %v", err)
		return err
	}
	return nil
}

func GetNginxConfig(ctx context.Context, m config.MachineConfig) <-chan struct{} {
	done := make(chan struct{})

	go func() {
		c := NewCollecter(m)
		defer c.Close()

		// time.Sleep(time.Second * 8)

		if err := c.Connect(); err != nil {
			log.Errorf("collecter connect faild: %v", err)
			return
		}

		if err := c.Fetch(); err != nil {
			log.Errorf("collecter fetch faild: %v", err)
			return
		}

		// signal done
		close(done)
		// done <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		log.Errorf("collecting %s-%s timeout", m.Alias, m.Host)
	case <-done:
		log.Infof("collecting %s-%s done", m.Alias, m.Host)
	}

	return done
}
