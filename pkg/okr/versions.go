package okr

import (
	"context"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/oneblock-ai/okr/pkg/kubectl"
)

func (o *OKR) getExistingVersions(ctx context.Context) (k8sVersion, kubeRayVersion string) {
	kubeConfig, err := kubectl.GetKubeconfig("")
	if err != nil {
		return "", ""
	}

	data, err := os.ReadFile(kubeConfig)
	if err != nil {
		return "", ""
	}

	restConfig, err := clientcmd.RESTConfigFromKubeConfig(data)
	if err != nil {
		return "", ""
	}

	k8s, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return "", ""
	}

	return getK8sVersion(ctx, k8s), getKubeRayVersion()
}

func getK8sVersion(ctx context.Context, k8s kubernetes.Interface) string {
	nodes, err := k8s.CoreV1().Nodes().List(ctx, metav1.ListOptions{
		LabelSelector: "node-role.kubernetes.io/control-plane=true",
	})
	if err != nil || len(nodes.Items) == 0 {
		return ""
	}
	return nodes.Items[0].Status.NodeInfo.KubeletVersion
}

func getKubeRayVersion() string {
	//TODO: add get KubeRay version by its deployment
	return ""
}
