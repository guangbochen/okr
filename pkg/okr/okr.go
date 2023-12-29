package okr

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"

	"github.com/oneblock-ai/okr/pkg/k3s/config"
	plan2 "github.com/oneblock-ai/okr/pkg/k3s/plan"
	"github.com/oneblock-ai/okr/pkg/k3s/versions"
)

const (
	// DefaultDataDir is the location of all state for okr
	DefaultDataDir = "/var/lib/oneblock-ai/okr"
	// DefaultConfigFile is the location of the okr config
	DefaultConfigFile = "/etc/oneblock-ai/okr/config.yaml"
)

type Config struct {
	Force      bool
	DataDir    string
	ConfigPath string
}

type UpgradeConfig struct {
	KubernetesVersion string
	KubeRayVersion    string
	Force             bool
}

type OKR struct {
	cfg Config
}

func New(cfg Config) *OKR {
	return &OKR{
		cfg: cfg,
	}
}

func (o *OKR) Run(ctx context.Context) error {
	if done, err := o.done(); err != nil {
		return fmt.Errorf("checking done stamp [%s]: %w", o.DoneStamp(), err)
	} else if done {
		logrus.Infof("System is already bootstrapped. To force the system to be bootstrapped again run with the --force flag")
		return nil
	}

	for {
		err := o.execute(ctx)
		if err == nil {
			return nil
		}
		logrus.Infof("failed to bootstrap system, will retry: %v", err)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(15 * time.Second):
		}
	}
}

func (o *OKR) execute(ctx context.Context) error {
	cfg, err := config.Load(o.cfg.ConfigPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if err := o.setWorking(cfg); err != nil {
		return fmt.Errorf("saving working config to %s: %w", o.WorkingStamp(), err)
	}

	if cfg.Role == "" {
		logrus.Warn("No role defined, skipping bootstrap")
		return nil
	}

	k8sVersion, err := versions.K8sVersion(cfg.KubernetesVersion)
	if err != nil {
		return err
	}

	logrus.Infof("Bootstrapping Kubernetes (%s)", k8sVersion)

	nodePlan, err := plan2.ToPlan(ctx, &cfg, o.cfg.DataDir)
	if err != nil {
		return fmt.Errorf("generating plan: %w", err)
	}

	if err := plan2.Run(ctx, &cfg, nodePlan, o.cfg.DataDir); err != nil {
		return fmt.Errorf("running plan: %w", err)
	}

	if err := o.setDone(cfg); err != nil {
		return err
	}

	logrus.Infof("Successfully Bootstrapped Kubernetes (%s)", k8sVersion)
	return nil
}

func (o *OKR) writeConfig(path string, cfg config.Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0600); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(path), err)
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	return err
}

func (o *OKR) setWorking(cfg config.Config) error {
	return o.writeConfig(o.WorkingStamp(), cfg)
}

func (o *OKR) setDone(cfg config.Config) error {
	return o.writeConfig(o.DoneStamp(), cfg)
}

func (o *OKR) done() (bool, error) {
	if o.cfg.Force {
		_ = os.Remove(o.DoneStamp())
		return false, nil
	}
	_, err := os.Stat(o.DoneStamp())
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (o *OKR) DoneStamp() string {
	return filepath.Join(o.cfg.DataDir, "bootstrapped")
}

func (o *OKR) WorkingStamp() string {
	return filepath.Join(o.cfg.DataDir, "working")
}
