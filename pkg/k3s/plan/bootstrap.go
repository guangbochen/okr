package plan

import (
	"context"
	"fmt"

	"github.com/rancher/system-agent/pkg/applyinator"

	config2 "github.com/oneblock-ai/okr/pkg/k3s/config"
	"github.com/oneblock-ai/okr/pkg/k3s/instructions/probe"
	"github.com/oneblock-ai/okr/pkg/k3s/instructions/resources"
	runtime2 "github.com/oneblock-ai/okr/pkg/k3s/instructions/runtime"
	"github.com/oneblock-ai/okr/pkg/k3s/kubectl"
	"github.com/oneblock-ai/okr/pkg/k3s/registry"
	"github.com/oneblock-ai/okr/pkg/k3s/versions"
)

type plan applyinator.Plan

func toInitPlan(config *config2.Config, dataDir string) (*applyinator.Plan, error) {
	if err := assignTokenIfUnset(config); err != nil {
		return nil, err
	}

	plan := plan{}
	if err := plan.addFiles(config, dataDir); err != nil {
		return nil, err
	}

	if err := plan.addInstructions(config, dataDir, true); err != nil {
		return nil, err
	}

	if err := plan.addProbes(config); err != nil {
		return nil, err
	}

	return (*applyinator.Plan)(&plan), nil
}
func toJoinPlan(cfg *config2.Config, dataDir string) (*applyinator.Plan, error) {
	if cfg.Server == "" {
		return nil, fmt.Errorf("server is required in config for all roles besides cluster-init")
	}
	if cfg.Token == "" {
		return nil, fmt.Errorf("token is required in config for all roles besides cluster-init")
	}

	plan := plan{}

	// add join plan files
	if err := plan.addJoinFiles(cfg, dataDir); err != nil {
		return nil, err
	}

	// add join instructions
	if err := plan.addInstructions(cfg, dataDir, false); err != nil {
		return nil, err
	}

	// add probes
	if err := plan.addProbesForJoin(cfg); err != nil {
		return nil, err
	}

	return (*applyinator.Plan)(&plan), nil
}

func ToPlan(ctx context.Context, config *config2.Config, dataDir string) (*applyinator.Plan, error) {
	newCfg := *config
	if newCfg.Role == "cluster-init" {
		return toInitPlan(&newCfg, dataDir)
	}
	return toJoinPlan(&newCfg, dataDir)
}

func (p *plan) addInstructions(cfg *config2.Config, dataDir string, addResource bool) error {
	k8sVersion, err := versions.K8sVersion(cfg.KubernetesVersion)
	if err != nil {
		return err
	}

	// add runtime instruction, e.g., k3s
	if err := p.addOneTimeInstruction(runtime2.ToInstruction(&cfg.RuntimeConfig, cfg.RuntimeInstallerImage, cfg.SystemDefaultRegistry, k8sVersion)); err != nil {
		return err
	}

	// add probe instruction
	if err := p.addOneTimeInstruction(probe.ToInstruction()); err != nil {
		return err
	}

	// add resource instruction
	if addResource {
		if err := p.addOneTimeInstruction(resources.ToInstruction(cfg.RuntimeInstallerImage, cfg.SystemDefaultRegistry, k8sVersion, dataDir)); err != nil {
			return err
		}
	}

	p.addPrePostInstructions(cfg, k8sVersion)
	return nil
}

func (p *plan) addPrePostInstructions(cfg *config2.Config, k8sVersion string) {
	var instructions []applyinator.OneTimeInstruction

	for _, inst := range cfg.PreOneTimeInstructions {
		if k8sVersion != "" {
			inst.Env = append(inst.Env, kubectl.Env(k8sVersion)...)
		}
		instructions = append(instructions, inst)
	}

	instructions = append(instructions, p.OneTimeInstructions...)

	for _, inst := range cfg.PostOneTimeInstructions {
		inst.Env = append(inst.Env, kubectl.Env(k8sVersion)...)
		instructions = append(instructions, inst)
	}

	p.OneTimeInstructions = instructions
	return
}

func (p *plan) addOneTimeInstruction(instruction *applyinator.OneTimeInstruction, err error) error {
	if err != nil || instruction == nil {
		return err
	}

	p.OneTimeInstructions = append(p.OneTimeInstructions, *instruction)
	return nil
}

func (p *plan) addPeriodInstruction(instruction *applyinator.PeriodicInstruction, err error) error {
	if err != nil || instruction == nil {
		return err
	}

	p.PeriodicInstructions = append(p.PeriodicInstructions, *instruction)
	return nil
}

// addFiles helps to generate plan files
func (p *plan) addFiles(cfg *config2.Config, dataDir string) error {
	k8sVersions, err := versions.K8sVersion(cfg.KubernetesVersion)
	if err != nil {
		return err
	}
	runtimeName := config2.GetRuntime(k8sVersions)

	// config.yaml
	if err := p.addFile(runtime2.ToFile(&cfg.RuntimeConfig, runtimeName, true)); err != nil {
		return err
	}

	// registries.yaml
	if err := p.addFile(registry.ToFile(cfg.Registries, runtimeName)); err != nil {
		return err
	}

	// bootstrap manifests
	if err := p.addFile(resources.ToBootstrapFile(cfg, resources.GetBootstrapManifests(dataDir))); err != nil {
		return err
	}

	return nil
}
func (p *plan) addJoinFiles(cfg *config2.Config, dataDir string) error {
	k8sVersions, err := versions.K8sVersion(cfg.KubernetesVersion)
	if err != nil {
		return err
	}
	runtimeName := config2.GetRuntime(k8sVersions)

	// config.yaml
	if err := p.addFile(runtime2.ToFile(&cfg.RuntimeConfig, runtimeName, false)); err != nil {
		return err
	}
	return nil
}

func (p *plan) addFile(file *applyinator.File, err error) error {
	if err != nil || file == nil {
		return err
	}
	p.Files = append(p.Files, *file)
	return nil
}

func (p *plan) addProbesForJoin(cfg *config2.Config) error {
	p.Probes = probe.ProbesForJoin(&cfg.RuntimeConfig)
	return nil
}

func (p *plan) addProbes(cfg *config2.Config) error {
	k8sVersion, err := versions.K8sVersion(cfg.KubernetesVersion)
	if err != nil {
		return err
	}
	p.Probes = probe.AllProbes(config2.GetRuntime(k8sVersion))
	return nil
}
