package buildkit

import (
	"fmt"
	"strings"

	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/util/system"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	p "github.com/railwayapp/railpack-go/core/plan"
)

type ConvertPlanOptions struct {
	BuildPlatform BuildPlatform
	CacheKey      string
}

const (
	WorkingDir = "/app"
)

func ConvertPlanToLLB(plan *p.BuildPlan, opts ConvertPlanOptions) (*llb.State, *Image, error) {
	platform := opts.BuildPlatform.ToPlatform()

	state := getBaseState(platform)

	// Add all variables as environment variables
	for name, value := range plan.Variables {
		state = state.AddEnv(name, value)
	}

	cacheStore := NewBuildKitCacheStore(opts.CacheKey)

	graph, err := NewBuildGraph(plan, &state, cacheStore)
	if err != nil {
		return nil, nil, err
	}

	graphOutput, err := graph.GenerateLLB()
	if err != nil {
		return nil, nil, err
	}

	state = *graphOutput.State
	imageEnv := getImageEnv(graphOutput)

	state = getStartState(state, plan, platform)

	startCommand := plan.Start.Command
	if plan.Start.Command == "" {
		startCommand = "/bin/bash"
	}

	image := Image{
		Image: specs.Image{
			Platform: specs.Platform{
				OS:           platform.OS,
				Architecture: platform.Architecture,
			},
		},
		Variant: platform.Variant,
		Config: specs.ImageConfig{
			Env:        imageEnv,
			WorkingDir: WorkingDir,
			Entrypoint: []string{"/bin/sh", "-c"},
			Cmd:        []string{startCommand},
		},
	}

	return &state, &image, nil
}

func getStartState(buildState llb.State, plan *p.BuildPlan, platform specs.Platform) llb.State {
	startState := buildState
	startState.Dir(WorkingDir)

	if plan.Start.BaseImage != "" {
		// This is all the user code + any modifications made by the providers
		mergedState := startState.File(llb.Copy(llb.Local("context"), ".", ".", &llb.CopyInfo{
			CreateDestPath:      true,
			FollowSymlinks:      true,
			CopyDirContentsOnly: true,
		}))

		// The base image to start from
		startState = llb.Image(plan.Start.BaseImage, llb.Platform(platform))

		// Copy over necessary files to the start image
		for _, path := range plan.Start.Paths {
			startState = startState.File(llb.Copy(mergedState, path, path, &llb.CopyInfo{
				CreateDestPath:      true,
				FollowSymlinks:      true,
				CopyDirContentsOnly: true,
			}))
		}
	} else {
		// If there is no custom start image, we will just copy over any additional files from the local context
		src := llb.Local("context")
		for _, path := range plan.Start.Paths {
			startState = startState.Dir(WorkingDir).File(llb.Copy(src, path, path, &llb.CopyInfo{
				CreateDestPath:      true,
				FollowSymlinks:      true,
				CopyDirContentsOnly: true,
			}))
		}
	}

	return startState
}

func getImageEnv(graphOutput *BuildGraphOutput) []string {
	pathString := strings.Join(graphOutput.PathList, ":")

	var pathEnv string
	if pathString == "" {
		pathEnv = "PATH=" + system.DefaultPathEnvUnix
	} else {
		pathEnv = "PATH=" + pathString + ":" + system.DefaultPathEnvUnix
	}

	imageEnv := []string{pathEnv}

	for k, v := range graphOutput.EnvVars {
		imageEnv = append(imageEnv, fmt.Sprintf("%s=%s", k, v))
	}

	return imageEnv
}

func getBaseState(platform specs.Platform) llb.State {
	state := llb.Image("ubuntu:noble",
		llb.Platform(platform),
	)

	state.Run(llb.Shlex("sh -c 'apt-get update && apt-get install -y --no-install-recommends git ca-certificates && rm -rf /var/lib/apt/lists/*'"), llb.WithCustomName("install base packages"))
	state.Dir(WorkingDir)

	return state
}
