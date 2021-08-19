package scalr

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModuleVersionsList(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()
	m, err := client.Modules.Read(ctx, defaultModuleID)
	require.NoError(t, err)
	assert.Equal(t, m.ID, defaultModuleID)

	t.Run("without module options", func(t *testing.T) {
		_, err := client.ModuleVersions.List(ctx, ModuleVersionListOptions{})
		require.Error(t, err)
		assert.EqualError(t, err, "filter[module] is required")
	})

	t.Run("with list options", func(t *testing.T) {
		ml, err := client.ModuleVersions.List(ctx, ModuleVersionListOptions{
			ListOptions: ListOptions{
				PageNumber: 999,
				PageSize:   100,
			},
			Module: defaultModuleID,
		})
		require.NoError(t, err)
		assert.Empty(t, ml.Items)
		assert.Equal(t, 999, ml.CurrentPage)
	})
}
