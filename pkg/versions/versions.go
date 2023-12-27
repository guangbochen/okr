package versions

import (
	"fmt"
	"net/http"
	"path"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

var (
	cachedK8sVersion = map[string]string{}
	cachedLock       sync.Mutex
	redirectClient   = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
)

func getVersionOrURL(urlFormat, def, version string) (_ string, isURL bool) {
	if version == "" {
		version = def
	}

	if strings.HasPrefix(version, "v") && len(strings.Split(version, ".")) > 2 {
		return version, false
	}

	channelURL := version
	if !strings.HasPrefix(channelURL, "https://") &&
		!strings.HasPrefix(channelURL, "http://") {
		if strings.HasSuffix(channelURL, "-head") || strings.Contains(channelURL, "/") {
			return channelURL, false
		}
		channelURL = fmt.Sprintf(urlFormat, version)
	}

	return channelURL, true
}

func K8sVersion(kubernetesVersion string) (string, error) {
	cachedLock.Lock()
	defer cachedLock.Unlock()

	cached, ok := cachedK8sVersion[kubernetesVersion]
	if ok {
		return cached, nil
	}

	urlFormat := "https://update.k3s.io/v1-release/channels/%s"
	if strings.HasSuffix(kubernetesVersion, ":k3s") {
		kubernetesVersion = strings.TrimSuffix(kubernetesVersion, ":k3s")
	}

	versionOrURL, isURL := getVersionOrURL(urlFormat, "stable", kubernetesVersion)
	if !isURL {
		return versionOrURL, nil
	}

	resp, err := redirectClient.Get(versionOrURL)
	if err != nil {
		return "", fmt.Errorf("getting channel version from (%s): %w", versionOrURL, err)
	}
	defer resp.Body.Close()

	url, err := resp.Location()
	if err != nil {
		return "", fmt.Errorf("getting channel version URL from (%s): %w", versionOrURL, err)
	}

	resolved := path.Base(url.Path)
	cachedK8sVersion[kubernetesVersion] = resolved
	logrus.Infof("Resolving Kubernetes version [%s] to %s from %s ", kubernetesVersion, resolved, versionOrURL)
	return resolved, nil
}
