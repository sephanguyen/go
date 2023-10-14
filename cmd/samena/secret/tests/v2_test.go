package tests

// We don't need these tests anymore, but we may need it in the future,
// when we do migration for platform services (nats/unleash/...).

// import (
// 	"fmt"
// 	"io/fs"
// 	"path/filepath"
// 	"strings"
// 	"testing"

// 	"github.com/manabie-com/backend/internal/golibs/configs"
// 	"github.com/manabie-com/backend/internal/golibs/execwrapper"
// 	"go.mozilla.org/sops/v3/decrypt"

// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// )

// // TestSecretV2 ensures v2 secrets are identical to v1 secrets.
// func TestSecretV2(t *testing.T) {
// 	// Only check in deployments/helm directory
// 	targetDir := filepath.Join(execwrapper.RootDirectory(), "deployments/helm")
// 	err := filepath.WalkDir(targetDir, func(path string, d fs.DirEntry, err error) error {
// 		require.NoError(t, err, "filepath.WalkDir failed")
// 		if d.IsDir() {
// 			return nil
// 		}
// 		if strings.Contains(path, "import-map-deployer") || strings.Contains(path, "yugabyte") || !strings.Contains(path, "secret") {
// 			return nil
// 		}
// 		if configs.IsSecretV2(path) {
// 			v1Path := getV1Path(path)
// 			v2, err := decrypt.File(path, "")
// 			require.NoErrorf(t, err, "failed to decrypt v2 secret at %q", path)
// 			v1, err := decrypt.File(v1Path, "")
// 			require.NoErrorf(t, err, "failed to decrypt v1 secret at %q", v1Path)
// 			assert.Equal(t, v1, v2, "v2 secret does not match v1 secret for %q", path)
// 		}
// 		return nil
// 	})
// 	require.NoError(t, err)
// }

// func getV1Path(v2Path string) string {
// 	v1Path := strings.ReplaceAll(v2Path, "_v2", "")
// 	v1Path = strings.ReplaceAll(v1Path, ".v2", "")
// 	if v1Path == v2Path {
// 		panic(fmt.Errorf("not a v2 secret path: %s", v2Path))
// 	}
// 	return v1Path
// }
