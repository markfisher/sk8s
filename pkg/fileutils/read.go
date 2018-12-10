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
 */

package fileutils

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
)

// Read reads the contents of the specified file. If the file is a relative path, it is relative to base.
// Either file or base may be URLs.
func Read(file string, base string) ([]byte, error) {
	absoluteFile, err := absFile(file, base)
	if err != nil {
		return nil, err
	}
	return readAbsFile(absoluteFile)
}

func absFile(file string, base string) (string, error) {
	u, err := url.Parse(file)
	if err != nil {
		return "", err
	}
	if u.IsAbs() {
		return u.String(), nil
	}

	if filepath.IsAbs(file) {
		return file, nil
	}

	b, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	if b.IsAbs() {
		b.Path = path.Join(b.Path, file)
		return b.String(), nil
	}

	if filepath.IsAbs(base) {
		return filepath.Join(base, file), nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(wd, base, file), nil
}

func readAbsFile(file string) ([]byte, error) {
	u, err := url.Parse(file)
	if err != nil {
		return nil, err
	}
	if u.IsAbs() {
		return downloadFile(u.String())
	}

	if !filepath.IsAbs(file) {
		panic(fmt.Sprintf("readAbsFile called with relative file path: %s", file))
	}

	return ioutil.ReadFile(file)
}

func downloadFile(url string) ([]byte, error) {
	t := &http.Transport{}
	t.RegisterProtocol("file", http.NewFileTransport(http.Dir("/")))
	c := &http.Client{Transport: t}
	resp, err := c.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
