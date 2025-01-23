package node

import (
	"testing"

	testingUtils "github.com/railwayapp/railpack-go/core/testing"
	"github.com/stretchr/testify/require"
)

func TestPackageManager(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		packageManager PackageManager
	}{
		{
			name:           "npm project",
			path:           "../../../examples/node-npm-latest",
			packageManager: PackageManagerNpm,
		},
		{
			name:           "bun project",
			path:           "../../../examples/node-bun",
			packageManager: PackageManagerBun,
		},
		{
			name:           "pnpm project",
			path:           "../../../examples/node-corepack",
			packageManager: PackageManagerPnpm,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, tt.path)
			provider := NodeProvider{}

			packageManager := provider.getPackageManager(ctx.App)
			require.Equal(t, tt.packageManager, packageManager)
		})
	}
}

func TestNodeCorepack(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		wantCorepack bool
	}{
		{
			name:         "corepack project",
			path:         "../../../examples/node-corepack",
			wantCorepack: true,
		},
		{
			name:         "bun project",
			path:         "../../../examples/node-bun",
			wantCorepack: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, tt.path)
			provider := NodeProvider{}

			packageJson, err := provider.getPackageJson(ctx.App)
			require.NoError(t, err)

			usesCorepack := provider.usesCorepack(packageJson)
			require.Equal(t, tt.wantCorepack, usesCorepack)
		})
	}
}
