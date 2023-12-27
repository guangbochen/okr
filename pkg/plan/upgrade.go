package plan

//import (
//	"github.com/oneblock-ai/okr/pkg/os"
//	"github.com/oneblock-ai/okr/pkg/rancher"
//	"github.com/rancher/system-agent/pkg/applyinator"
//
//	"github.com/oneblock-ai/okr/pkg/config"
//	"github.com/oneblock-ai/okr/pkg/runtime"
//)
//
//func Upgrade(cfg *config.Config, k8sVersion, rancherVersion, rancherOSVersion, dataDir string) (*applyinator.Plan, error) {
//	p := plan{}
//
//	if rancherVersion != "" {
//		if err := p.addInstruction(rancher.ToUpgradeInstruction("", cfg.SystemDefaultRegistry, k8sVersion, rancherVersion, dataDir)); err != nil {
//			return nil, err
//		}
//		if err := p.addInstruction(rancher.ToWaitRancherInstruction("", cfg.SystemDefaultRegistry, k8sVersion)); err != nil {
//			return nil, err
//		}
//	}
//
//	if k8sVersion != "" {
//		if err := p.addInstruction(runtime.ToUpgradeInstruction(k8sVersion)); err != nil {
//			return nil, err
//		}
//		if err := p.addInstruction(runtime.ToWaitKubernetesInstruction("", cfg.SystemDefaultRegistry, k8sVersion)); err != nil {
//			return nil, err
//		}
//	}
//
//	if rancherOSVersion != "" {
//		if err := p.addInstruction(os.ToUpgradeInstruction(k8sVersion, rancherOSVersion)); err != nil {
//			return nil, err
//		}
//	}
//
//	return (*applyinator.Plan)(&p), nil
//}
