package pack

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/buildpack/lifecycle"
	"github.com/buildpack/packs"
	"github.com/buildpack/packs/img"
	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	dockercli "github.com/docker/docker/client"
	"github.com/pkg/errors"
)

func exportRegistry(group *lifecycle.BuildpackGroup, workspaceDir, repoName, stackName string) (string, error) {
	origImage, err := readImage(repoName, false)
	if err != nil {
		return "", err
	}

	stackImage, err := readImage(stackName, false)
	if err != nil || stackImage == nil {
		return "", packs.FailErr(err, "get image for", stackName)
	}

	repoStore, err := img.NewRegistry(repoName)
	if err != nil {
		return "", packs.FailErr(err, "access", repoName)
	}

	tmpDir, err := ioutil.TempDir("", "lifecycle.exporter.layer")
	if err != nil {
		return "", packs.FailErr(err, "create temp directory")
	}
	defer os.RemoveAll(tmpDir)

	exporter := &lifecycle.Exporter{
		Buildpacks: group.Buildpacks,
		TmpDir:     tmpDir,
		Out:        os.Stdout,
		Err:        os.Stderr,
	}
	newImage, err := exporter.Export(
		workspaceDir,
		stackImage,
		origImage,
	)
	if err != nil {
		return "", packs.FailErrCode(err, packs.CodeFailedBuild)
	}

	if err := repoStore.Write(newImage); err != nil {
		return "", packs.FailErrCode(err, packs.CodeFailedUpdate, "write")
	}

	sha, err := newImage.Digest()
	if err != nil {
		return "", packs.FailErr(err, "calculating image digest")
	}

	return sha.String(), nil
}

func exportDaemon(buildpacks []string, workspaceVolume, repoName, runImage string) error {
	cli, err := dockercli.NewEnvClient()
	if err != nil {
		return errors.Wrap(err, "new docker client")
	}
	ctx := context.Background()
	ctr, err := cli.ContainerCreate(ctx, &container.Config{
		Image:      runImage,
		User:       "root",
		Entrypoint: []string{},
		Cmd:        []string{"echo", "hi"},
	}, &container.HostConfig{
		Binds: []string{
			workspaceVolume + ":/workspace",
		},
	}, nil, "")
	if err != nil {
		return errors.Wrap(err, "container create")
	}

	r, _, err := cli.CopyFromContainer(ctx, ctr.ID, "/workspace")
	if err != nil {
		return errors.Wrap(err, "copy from container")
	}

	r2, layerChan, errChan := addDockerfileToTar(runImage, repoName, buildpacks, r)

	res, err := cli.ImageBuild(ctx, r2, dockertypes.ImageBuildOptions{Tags: []string{repoName}})
	if err != nil {
		return errors.Wrap(err, "image build")
	}
	defer res.Body.Close()
	if _, err := parseImageBuildBody(res.Body, os.Stdout); err != nil {
		return errors.Wrap(err, "image build")
	}
	res.Body.Close()

	if err := <-errChan; err != nil {
		return errors.Wrap(err, "modify tar to add dockerfile")
	}
	layerNames := <-layerChan

	// Calculate metadata
	i, _, err := cli.ImageInspectWithRaw(ctx, repoName)
	if err != nil {
		return errors.Wrap(err, "inspect image to find layers")
	}
	layerIDX := len(i.RootFS.Layers) - len(layerNames)
	metadata := packs.BuildMetadata{
		RunImage: packs.RunImageMetadata{
			Name: runImage,
			SHA:  i.RootFS.Layers[layerIDX-3],
		},
		App: packs.AppMetadata{
			SHA: i.RootFS.Layers[layerIDX-2],
		},
		Config: packs.ConfigMetadata{
			SHA: i.RootFS.Layers[layerIDX-1],
		},
		Buildpacks: []packs.BuildpackMetadata{},
	}
	for _, buildpack := range buildpacks {
		data := packs.BuildpackMetadata{Key: buildpack, Layers: make(map[string]packs.LayerMetadata)}
		for _, layer := range layerNames {
			if layer.buildpack == buildpack {
				data.Layers[layer.layer] = packs.LayerMetadata{
					SHA:  i.RootFS.Layers[layerIDX],
					Data: layer.data,
				}
				layerIDX++
			}
		}
		metadata.Buildpacks = append(metadata.Buildpacks, data)
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return errors.Wrap(err, "marshal metadata to json")
	}
	if err := addLabelToImage(cli, repoName, map[string]string{"sh.packs.build": string(metadataJSON)}); err != nil {
		return errors.Wrap(err, "add sh.packs.build label to image")
	}

	return nil
}

func addLabelToImage(cli *dockercli.Client, repoName string, labels map[string]string) error {
	dockerfile := "FROM " + repoName + "\n"
	for k, v := range labels {
		dockerfile += fmt.Sprintf("LABEL %s='%s'\n", k, v)
	}
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	tw.WriteHeader(&tar.Header{Name: "Dockerfile", Size: int64(len(dockerfile)), Mode: 0666})
	tw.Write([]byte(dockerfile))
	tw.Close()

	res, err := cli.ImageBuild(context.Background(), bytes.NewReader(buf.Bytes()), dockertypes.ImageBuildOptions{
		Tags: []string{repoName},
	})
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if _, err := parseImageBuildBody(res.Body, os.Stdout); err != nil {
		return errors.Wrap(err, "image build")
	}
	return err
}

type dockerfileLayer struct {
	buildpack string
	layer     string
	data      interface{}
}

func addDockerfileToTar(runImage, repoName string, buildpacks []string, r io.Reader) (io.Reader, chan []dockerfileLayer, chan error) {
	dockerFile := "FROM " + runImage + "\n"
	dockerFile += "ADD --chown=pack:pack /workspace/app /workspace/app\n"
	dockerFile += "ADD --chown=pack:pack /workspace/config /workspace/config\n"
	layerChan := make(chan []dockerfileLayer, 1)
	var layerNames []dockerfileLayer
	errChan := make(chan error, 1)

	pr, pw := io.Pipe()
	tw := tar.NewWriter(pw)

	isBuildpack := make(map[string]bool)
	for _, b := range buildpacks {
		isBuildpack[b] = true
	}

	go func() {
		defer pw.Close()
		tr := tar.NewReader(r)
		tomlFiles := make(map[string]map[string]interface{})
		dirs := make(map[string]map[string]bool)
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break // End of archive
			}
			if err != nil {
				layerChan <- nil
				errChan <- errors.Wrap(err, "tr.Next")
				return
			}

			tw.WriteHeader(hdr)

			arr := strings.Split(strings.TrimPrefix(strings.TrimSuffix(hdr.Name, "/"), "/"), "/")
			if len(arr) == 3 && isBuildpack[arr[1]] && strings.HasSuffix(arr[2], ".toml") && arr[2] != "launch.toml" {
				if tomlFiles[arr[1]] == nil {
					tomlFiles[arr[1]] = make(map[string]interface{})
				}

				buf, err := ioutil.ReadAll(tr)
				if err != nil {
					layerChan <- nil
					errChan <- errors.Wrap(err, "read toml file")
					return
				}
				if _, err := tw.Write(buf); err != nil {
					layerChan <- nil
					errChan <- errors.Wrap(err, "write toml file")
					return
				}

				var data interface{}
				if _, err := toml.Decode(string(buf), &data); err != nil {
					layerChan <- nil
					errChan <- errors.Wrap(err, "parsing toml file: "+hdr.Name)
					return
				}
				tomlFiles[arr[1]][strings.TrimSuffix(arr[2], ".toml")] = data
			} else if len(arr) == 3 && isBuildpack[arr[1]] && hdr.Typeflag == tar.TypeDir {
				if dirs[arr[1]] == nil {
					dirs[arr[1]] = make(map[string]bool)
				}
				dirs[arr[1]][arr[2]] = true
			}

			// TODO is it OK to do this if we have already read it above? eg. toml file
			if _, err := io.Copy(tw, tr); err != nil {
				layerChan <- nil
				errChan <- errors.Wrap(err, "io copy")
				return
			}
		}

		copyFromPrev := false
		for _, buildpack := range buildpacks {
			layers := sortedKeys(tomlFiles[buildpack])
			for _, layer := range layers {
				layerNames = append(layerNames, dockerfileLayer{buildpack, layer, tomlFiles[buildpack][layer]})
				if dirs[buildpack][layer] {
					dockerFile += fmt.Sprintf("ADD --chown=pack:pack /workspace/%s/%s /workspace/%s/%s\n", buildpack, layer, buildpack, layer)
				} else {
					dockerFile += fmt.Sprintf("COPY --from=prev --chown=pack:pack /workspace/%s/%s /workspace/%s/%s\n", buildpack, layer, buildpack, layer)
					copyFromPrev = true
				}
			}
		}
		if copyFromPrev {
			dockerFile = "FROM " + repoName + " AS prev\n\n" + dockerFile
		}

		if err := tw.WriteHeader(&tar.Header{Name: "Dockerfile", Size: int64(len(dockerFile)), Mode: 0666}); err != nil {
			layerChan <- nil
			errChan <- errors.Wrap(err, "write tar header for Dockerfile")
			return
		}
		if _, err := tw.Write([]byte(dockerFile)); err != nil {
			layerChan <- nil
			errChan <- errors.Wrap(err, "write Dockerfile to tar")
			return
		}

		tw.Close()
		pw.Close()
		layerChan <- layerNames
		errChan <- nil
	}()

	return pr, layerChan, errChan
}

func sortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for key, _ := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func parseImageBuildBody(r io.Reader, out io.Writer) (string, error) {
	jr := json.NewDecoder(r)
	var id string
	var streamError error
	var obj struct {
		Stream string `json:"stream"`
		Error  string `json:"error"`
		Aux    struct {
			ID string `json:"ID"`
		} `json:"aux"`
	}
	for {
		err := jr.Decode(&obj)
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
		if obj.Aux.ID != "" {
			id = obj.Aux.ID
		}
		if txt := strings.TrimSpace(obj.Stream); txt != "" {
			fmt.Fprintln(out, txt)
		}
		if txt := strings.TrimSpace(obj.Error); txt != "" {
			streamError = errors.New(txt)
		}
	}
	return id, streamError
}
