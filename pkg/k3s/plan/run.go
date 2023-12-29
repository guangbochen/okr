package plan

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/rancher/system-agent/pkg/applyinator"
	"github.com/rancher/system-agent/pkg/image"
	"github.com/sirupsen/logrus"

	config2 "github.com/oneblock-ai/okr/pkg/k3s/config"
	"github.com/oneblock-ai/okr/pkg/k3s/registry"
	"github.com/oneblock-ai/okr/pkg/k3s/versions"
)

func Run(ctx context.Context, cfg *config2.Config, plan *applyinator.Plan, dataDir string) error {
	k8sVersion, err := versions.K8sVersion(cfg.KubernetesVersion)
	if err != nil {
		return err
	}
	return RunWithKubernetesVersion(ctx, k8sVersion, plan, dataDir)
}

func RunWithKubernetesVersion(ctx context.Context, k8sVersion string, plan *applyinator.Plan, dataDir string) error {
	runtime := config2.GetRuntime(k8sVersion)

	if err := writePlan(plan, dataDir); err != nil {
		return err
	}

	images := image.NewUtility("", "", "", registry.GetConfigFile(runtime))
	apply := applyinator.NewApplyinator(filepath.Join(dataDir, "plan", "work"), false,
		filepath.Join(dataDir, "plan", "applied"), "", images)

	applyOutput, err := apply.Apply(ctx, applyinator.ApplyInput{
		CalculatedPlan: applyinator.CalculatedPlan{
			Plan: *plan,
		},
		RunOneTimeInstructions:     true,
		OneTimeInstructionAttempts: 5,
		ReconcileFiles:             true,
	})
	if err != nil {
		return err
	}

	return saveOutput(applyOutput.OneTimeOutput, dataDir)
}

func saveOutput(data []byte, dataDir string) error {
	planOutput := GetPlanOutput(dataDir)
	f, err := os.OpenFile(planOutput, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	in, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	_, err = io.Copy(f, in)
	return err
}

func writePlan(plan *applyinator.Plan, dataDir string) error {
	planFile := GetPlanFile(dataDir)
	if err := os.MkdirAll(filepath.Dir(planFile), 0755); err != nil {
		return err
	}

	logrus.Infof("Writing plan file to %s", planFile)
	f, err := os.OpenFile(planFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(plan)
}

func GetPlanFile(dataDir string) string {
	return filepath.Join(dataDir, "plan", "plan.json")
}

func GetPlanOutput(dataDir string) string {
	return filepath.Join(dataDir, "plan", "plan-output.json")
}
