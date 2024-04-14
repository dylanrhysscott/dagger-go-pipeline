// A dagger module for building and pushing Go based programs

package main

import (
	"context"
	"fmt"
)

type GoPipeline struct {
	hasAuth          bool
	registryHost     string
	registryUsername *Secret
	registryPassword *Secret
}

// Defines registry auth parameters
func (m *GoPipeline) WithRegistry(ctx context.Context, registryHost string, registryUsername *Secret, registryPassword *Secret) (*GoPipeline, error) {
	m.hasAuth = true
	m.registryHost = registryHost
	m.registryUsername = registryUsername
	m.registryPassword = registryPassword
	return m, nil
}

// Build takes go source in the given context and builds it into a container
func (m *GoPipeline) Build(
	ctx context.Context,
	name string,
	buildContext Directory,
	// +optional
	// +default="golang:1.22.1"
	buildImage string,
	// +optional
	// +default="alpine:latest"
	runImage string,
	// +optional
	// +default="3000"
	port int,
) (*Container, error) {
	srcPath := "/src"
	binFile := fmt.Sprintf("%s/%s", srcPath, name)
	entrypoint := fmt.Sprintf("/%s", name)
	build := dag.Container().
		From(buildImage).
		WithDirectory(srcPath, &buildContext).
		WithWorkdir(srcPath).
		WithEnvVariable("CGO_ENABLED", "0").
		WithExec([]string{"go", "build", "-o", binFile})
	run := dag.
		Container().
		From(runImage).
		WithFile("/", build.File(binFile)).
		WithEntrypoint([]string{entrypoint}).
		WithExposedPort(port)

	if m.hasAuth {
		username, err := m.registryUsername.Plaintext(ctx)
		if err != nil {
			return nil, fmt.Errorf("could not read registry username: %s", err)
		}
		run = run.WithRegistryAuth(m.registryHost, username, m.registryPassword)
	}
	return run, nil
}
