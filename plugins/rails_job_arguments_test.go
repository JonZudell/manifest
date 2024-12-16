package plugins

import (
	"strings"
	"testing"

	"github.com/blakewilliams/customs"
	"github.com/stretchr/testify/require"
)

var argumentChangeDiff = `
diff --git a/app/jobs/greeter_job.rb b/app/jobs/greeter_job.rb
index abc1234..def5678 100644
--- a/app/jobs/greeter_job.rb
+++ b/app/jobs/greeter_job.rb
@@ -1,7 +1,7 @@
 class GreeterJob < ApplicationJob
   queue_as :default

-  def perform
+  def perform(name)
     # Job logic here
   end
 end`

func TestCustoms_NewFile(t *testing.T) {
	diff, err := customs.NewDiff(strings.NewReader(argumentChangeDiff))
	require.NoError(t, err)

	entry := &customs.Entry{Diff: diff}
	result := &customs.Result{Comments: make([]customs.Comment, 0)}

	err = RailsJobArguments(entry, result)
	require.NoError(t, err)

	require.Len(t, result.Comments, 1)
	comment := result.Comments[0]

	require.Equal(t, "app/jobs/greeter_job.rb", comment.File)
	require.Equal(t, uint(4), comment.Line)
	require.Equal(t, customs.SeverityWarn, comment.Severity)
}
