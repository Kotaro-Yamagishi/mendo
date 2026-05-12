package menu_test

import (
	"context"
	"errors"
	"testing"

	appmenu "mendo/internal/application/command/menu"
	importjob "mendo/internal/domain/import"
	"mendo/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ImportMenusUsecase_正常系(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		rows          []importjob.ImportRow
		wantTotalRows int
	}{
		{
			name: "複数行",
			rows: []importjob.ImportRow{
				{MenuName: "醤油ラーメン", Price: 800},
				{MenuName: "味噌ラーメン", Price: 900},
			},
			wantTotalRows: 2,
		},
		{
			name:          "空スライス",
			rows:          []importjob.ImportRow{},
			wantTotalRows: 0,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			jw := &testutil.StubJobWriter{}
			je := &testutil.SpyJobEnqueuer{}
			uc := appmenu.NewImportMenusUsecase(jw, je)

			jobID, err := uc.Execute(context.Background(), tc.rows)

			require.NoError(t, err)
			assert.NotEmpty(t, jobID)
			// ジョブが保存された
			require.Len(t, jw.SavedJobs, 1)
			assert.Equal(t, tc.wantTotalRows, jw.SavedJobs[0].TotalRows)
			// ワーカーにエンキューされた
			require.Len(t, je.EnqueuedJobs, 1)
			assert.Equal(t, jobID, string(je.EnqueuedJobs[0].ID))
		})
	}
}

func Test_ImportMenusUsecase_JobWriter失敗(t *testing.T) {
	t.Parallel()

	jw := &testutil.StubJobWriter{SaveErr: errors.New("db error")}
	je := &testutil.SpyJobEnqueuer{}
	uc := appmenu.NewImportMenusUsecase(jw, je)

	_, err := uc.Execute(context.Background(), []importjob.ImportRow{
		{MenuName: "ラーメン", Price: 700},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
	// Save に失敗したらエンキューされない
	assert.Empty(t, je.EnqueuedJobs)
}
