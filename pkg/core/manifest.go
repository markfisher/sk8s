/*
 * Copyright 2018 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package core

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
)

const manifestVersion_0_1 = "0.1"

var manifests = map[string]*Manifest{
	"latest": &Manifest{
		ManifestVersion: manifestVersion_0_1,
		Istio: []string{
			"https://storage.googleapis.com/knative-releases/serving/latest/istio.yaml",
		},
		Knative: []string{
			"https://storage.googleapis.com/knative-releases/serving/latest/release-no-mon.yaml",
			"https://storage.googleapis.com/knative-releases/eventing/latest/release.yaml",
			"https://storage.googleapis.com/knative-releases/eventing/latest/release-clusterbus-stub.yaml",
		},
		Namespace: []string{
			"https://storage.googleapis.com/riff-releases/previous/riff-build/riff-build-0.1.0.yaml",
		},
	},
	"stable": &Manifest{
		ManifestVersion: manifestVersion_0_1,
		Istio: []string{
			"https://storage.googleapis.com/knative-releases/serving/previous/v20180828-7c20145/istio.yaml",
		},
		Knative: []string{
			"https://storage.googleapis.com/knative-releases/serving/previous/v20180828-7c20145/release-no-mon.yaml",
			"https://storage.googleapis.com/knative-releases/eventing/previous/v20180830-5d35af5/release.yaml",
			"https://storage.googleapis.com/knative-releases/eventing/previous/v20180830-5d35af5/release-clusterbus-stub.yaml",
		},
		Namespace: []string{
			"https://storage.googleapis.com/riff-releases/previous/riff-build/riff-build-0.1.0.yaml",
		},
	},
	"v0.1.2": &Manifest{
		ManifestVersion: manifestVersion_0_1,
		Istio: []string{
			"https://storage.googleapis.com/knative-releases/serving/previous/v20180828-7c20145/istio.yaml",
		},
		Knative: []string{
			"https://storage.googleapis.com/knative-releases/serving/previous/v20180828-7c20145/release-no-mon.yaml",
			"https://storage.googleapis.com/knative-releases/eventing/previous/v20180830-5d35af5/release.yaml",
			"https://storage.googleapis.com/knative-releases/eventing/previous/v20180830-5d35af5/release-clusterbus-stub.yaml",
		},
		Namespace: []string{
			"https://storage.googleapis.com/riff-releases/previous/riff-build/riff-build-0.1.0.yaml",
		},
	},
	"v0.1.1": &Manifest{
		ManifestVersion: manifestVersion_0_1,
		Istio: []string{
			"https://storage.googleapis.com/riff-releases/istio/istio-1.0.0-riff-crds.yaml",
			"https://storage.googleapis.com/riff-releases/istio/istio-1.0.0-riff-main.yaml",
		},
		Knative: []string{
			"https://storage.googleapis.com/knative-releases/serving/previous/v20180809-6b01d8e/release-no-mon.yaml",
			"https://storage.googleapis.com/knative-releases/eventing/previous/v20180809-34ab480/release.yaml",
			"https://storage.googleapis.com/knative-releases/eventing/previous/v20180809-34ab480/release-clusterbus-stub.yaml",
		},
		Namespace: []string{
			"https://storage.googleapis.com/riff-releases/previous/riff-build/riff-build-0.1.0.yaml",
		},
	},
	"v0.1.0": &Manifest{
		ManifestVersion: manifestVersion_0_1,
		Istio: []string{
			"https://storage.googleapis.com/riff-releases/istio-riff-0.1.0.yaml",
		},
		Knative: []string{
			"https://storage.googleapis.com/riff-releases/release-no-mon-riff-0.1.0.yaml",
			"https://storage.googleapis.com/riff-releases/release-eventing-riff-0.1.0.yaml",
			"https://storage.googleapis.com/riff-releases/release-eventing-clusterbus-stub-riff-0.1.0.yaml",
		},
		Namespace: []string{
			"https://storage.googleapis.com/riff-releases/previous/riff-build/riff-build-0.1.0.yaml",
		},
	},
}

// Manifest defines the location of YAML files for system components.
type Manifest struct {
	ManifestVersion string   `json:"manifestVersion"`
	Istio           []string `json:"istio"`
	Knative         []string `json:"knative"`
	Namespace       []string `json:"namespace"`
}

func NewManifest(path string) (*Manifest, error) {
	if manifest, ok := manifests[path]; ok {
		return manifest, nil
	}

	var m Manifest
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Error reading manifest file: %v", err)
	}

	err = yaml.Unmarshal(yamlFile, &m)
	if err != nil {
		return nil, fmt.Errorf("Error parsing manifest file: %v", err)
	}

	if m.ManifestVersion != manifestVersion_0_1 {
		return nil, fmt.Errorf("Manifest has unsupported version: %s", m.ManifestVersion)
	}

	err = checkCompleteness(m)
	if err != nil {
		return nil, err
	}

	err = convertManifestFilePathsToURLs(&m, filepath.Dir(path))
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func checkCompleteness(m Manifest) error {
	var omission string
	if m.Istio == nil {
		omission = "istio"
	} else if m.Knative == nil {
		omission = "knative"
	} else if m.Namespace == nil {
		omission = "namespace"
	} else {
		return nil
	}
	return fmt.Errorf("Manifest is incomplete: %s array missing: %#v", omission, m)
}

func convertManifestFilePathsToURLs(m *Manifest, baseDir string) error {
	for _, r := range []*[]string{&m.Istio, &m.Knative, &m.Namespace} {
		i, err := convertFilePathsToURLs(*r, baseDir)
		if err != nil {
			return err
		}
		*r = i
	}
	return nil
}

func convertFilePathsToURLs(paths []string, baseDir string) ([]string, error) {
	urls := []string{}
	for _, path := range paths {
		url, err := convertFilePathToURL(path, baseDir)
		if err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}
	return urls, nil
}

func convertFilePathToURL(path string, baseDir string) (string, error) {
	u, err := url.Parse(path)
	if err != nil {
		return "", err
	}
	if u.Scheme == "" {
		u.Scheme = "file"
		if !filepath.IsAbs(u.Path) {
			if !filepath.IsAbs(baseDir) {
				wd, err := os.Getwd()
				if err != nil {
					return "", err
				}
				baseDir = filepath.Join(wd, baseDir)
			}
			u.Path = fmt.Sprintf("%s/%s", baseDir, u.Path)
		}
		return u.String(), nil
	}
	return path, nil
}
