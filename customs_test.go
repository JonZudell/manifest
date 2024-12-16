package customs

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var newFile = `
diff --git a/README.md b/README.md
new file mode 100644
index 0000000..e69de29
--- /dev/null
+++ b/README.md
@@ -0,0 +1,1 @@
+# The truth is out there`

func TestCustoms_NewFile(t *testing.T) {
	diff, err := NewDiff(strings.NewReader(newFile))
	require.NoError(t, err)

	require.Len(t, diff.NewFiles, 1)

	readme := diff.Files["README.md"]
	require.Equal(t, DiffOperationNew, readme.Operation, "expected README to be new")

	require.Len(t, readme.Left, 0, "expected left side of README diff to be empty")
	require.Len(t, readme.Right, 1, "expected left side of README diff to be empty")

	line := readme.Right[0]
	require.Equal(t, "# The truth is out there", line.Content)
	require.Equal(t, uint(1), line.LineNo)
}
