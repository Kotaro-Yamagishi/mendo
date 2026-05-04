package importworker

import (
	"context"
	"fmt"
	"sync"

	importjob "mendo/internal/domain/import"
	menudomain "mendo/internal/domain/menu"
)

const chunkSize = 10 // チャンクサイズ（並列処理の単位）

// Worker は非同期でインポートジョブを処理する。
// チャンク分割 + goroutine で並列処理。
type Worker struct {
	jobStore   *InMemoryJobStore
	menuReader menudomain.Reader
	menuWriter menudomain.Writer
	jobCh      chan *importjob.Job
}

// NewWorker は Worker を作成する。
func NewWorker(
	jobStore *InMemoryJobStore,
	menuReader menudomain.Reader,
	menuWriter menudomain.Writer,
) *Worker {
	return &Worker{
		jobStore:   jobStore,
		menuReader: menuReader,
		menuWriter: menuWriter,
		jobCh:      make(chan *importjob.Job, 100),
	}
}

// Start はワーカーを起動する。バックグラウンドで goroutine が処理する。
func (w *Worker) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case job := <-w.jobCh:
				w.processJob(ctx, job)
			}
		}
	}()
	fmt.Println("[ImportWorker] started")
}

// Enqueue はジョブをキューに投入する。
func (w *Worker) Enqueue(job *importjob.Job) {
	w.jobCh <- job
}

func (w *Worker) processJob(ctx context.Context, job *importjob.Job) {
	fmt.Printf("[ImportWorker] processing job %s (%d rows)\n", job.ID, job.TotalRows)
	job.Start()

	// 既存メニュー名を一括取得（外部アクセス1回）
	existingMenus, err := w.menuReader.FindAll(ctx)
	if err != nil {
		fmt.Printf("[ImportWorker] failed to load menus: %v\n", err)
		job.Complete()
		return
	}
	existingNames := make(map[string]bool)
	for _, m := range existingMenus {
		existingNames[m.Name().String()] = true
	}

	// チャンク分割して並列処理
	rows := job.Rows
	var wg sync.WaitGroup
	for i := 0; i < len(rows); i += chunkSize {
		end := i + chunkSize
		if end > len(rows) {
			end = len(rows)
		}
		chunk := rows[i:end]
		offset := i

		wg.Add(1)
		go func(c []importjob.ImportRow, o int) {
			defer wg.Done()
			w.processChunk(ctx, job, c, o, existingNames)
		}(chunk, offset)
	}
	wg.Wait()

	job.Complete()
	fmt.Printf("[ImportWorker] job %s completed: %s\n", job.ID, job.Summary())
}

func (w *Worker) processChunk(
	ctx context.Context,
	job *importjob.Job,
	chunk []importjob.ImportRow,
	offset int,
	existingNames map[string]bool,
) {
	for i, row := range chunk {
		rowNum := offset + i + 1

		// 業務ルール: 同名メニュー重複チェック（インメモリ）
		if existingNames[row.MenuName] {
			job.RecordFailure(rowNum, row.MenuName, "同名メニューが既に存在します")
			continue
		}

		// 値オブジェクト生成（バリデーション）
		name, err := menudomain.NewMenuName(row.MenuName)
		if err != nil {
			job.RecordFailure(rowNum, row.MenuName, err.Error())
			continue
		}
		price, err := menudomain.NewPrice(row.Price)
		if err != nil {
			job.RecordFailure(rowNum, row.MenuName, err.Error())
			continue
		}

		// Menu 集約を生成して保存
		menuID := menudomain.MenuID(fmt.Sprintf("menu-%d-%d", offset, i))
		m := menudomain.NewMenu(menuID, name, price)
		if err := w.menuWriter.Save(ctx, m); err != nil {
			job.RecordFailure(rowNum, row.MenuName, fmt.Sprintf("保存失敗: %v", err))
			continue
		}

		existingNames[row.MenuName] = true // 同じバッチ内の重複も防ぐ
		job.RecordSuccess()
	}
}
