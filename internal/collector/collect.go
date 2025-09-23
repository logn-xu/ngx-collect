package collector

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kr/fs"
	"github.com/logn/ngx-collect/internal/config"
	"github.com/logn/ngx-collect/internal/sshclient"
	"github.com/pkg/sftp"
	log "github.com/sirupsen/logrus"
)

type Collecter interface {
	Connect() error
	Fetch() error
	Close() error
}

type CollecterImpl struct {
	config        config.MachineConfig
	client        *sftp.Client
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

	client, err := sftp.NewClient(sshClient)
	if err != nil {
		return err
	}

	c.client = client
	return nil
}

// fetch nginx files
func (c *CollecterImpl) Fetch() error {
	// create local base path
	if err := c.CreateBaseDestDir(); err != nil {
		log.Errorf("create base dir failed: %v", err)
		return err
	}

	// for each file in remote path
	for _, remotePath := range c.config.RemotePaths {
		// enter sub dir
		fs := c.client.Walk(remotePath)
		for fs.Step() {
			if fs.Err() != nil {
				log.Errorf("walk failed: %v", fs.Err())
				continue
			}

			// log.Infof(fs.Path())
			f, err := c.client.Stat(fs.Path())
			if err != nil {
				log.Errorf("stat failed: %v", err)
				continue
			}

			switch f.IsDir() {
			case true:
				if err := c.downloadDir(fs, remotePath); err != nil {
					log.Errorf("download file failed: %v", err)
				}
			case false:
				if err := c.downloadFile(fs, remotePath); err != nil {
					log.Errorf("download file failed: %v", err)
				}
			}
		}
	}

	return nil
}

func (c *CollecterImpl) Close() error {
	return c.client.Close()
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

func (c *CollecterImpl) downloadFile(fs *fs.Walker, remotePath string) error {
	localPath := generereLocalPath(fs.Path(), remotePath)
	// get remote file
	srcFile, err := c.client.OpenFile(fs.Path(), os.O_RDONLY)
	if err != nil {
		log.Errorf("open file failed: %v", err)
		return err
	}
	defer srcFile.Close()

	// get remote file info
	info, err := srcFile.Stat()
	if err != nil {
		log.Errorf("stat remote file failed: %v", err)
		return err
	}

	// create local file
	localFile, err := os.Create(c.localBasePath + "/" + localPath)
	if err != nil {
		log.Errorf("create local file failed: %v", err)
		return err
	}
	defer localFile.Close()
	// copy file
	if _, err := io.Copy(localFile, srcFile); err != nil {
		log.Errorf("copy file failed: %v", err)
		return err
	}

	// set file permission
	if err := os.Chmod(localFile.Name(), info.Mode()); err != nil {
		log.Errorf("chmod local file failed: %v", err)
		return err
	}

	return nil
}

func (c *CollecterImpl) downloadDir(fs *fs.Walker, remotePath string) error {
	localPath := generereLocalPath(fs.Path(), remotePath)
	if err := os.MkdirAll(fmt.Sprintf("%s/%s", c.localBasePath, localPath), 0755); err != nil {
		log.Errorf("create local dir failed: %v", err)
		return err
	}
	return nil
}

func generereLocalPath(fsPath string, remotePath string) string {
	fullRemotePath := filepath.Clean(remotePath)
	dir := filepath.Dir(fullRemotePath)
	return strings.Replace(fsPath, dir, "", 1)
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
