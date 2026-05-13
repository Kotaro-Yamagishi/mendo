package importworker

import (
	"context"
	"fmt"
	"log/slog"
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
	logger     *slog.Logger
}

// NewWorker は Worker を作成する。
func NewWorker(
	jobStore *InMemoryJobStore,
	menuReader menudomain.Reader,
	menuWriter menudomain.Writer,
	logger *slog.Logger,
) *Worker {
	return &Worker{
		jobStore:   jobStore,
		menuReader: menuReader,
		menuWriter: menuWriter,
		jobCh:      make(chan *importjob.Job, 100),
		logger:     logger,
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
	w.logger.Info("import worker started")
}

// Enqueue はジョブをキューに投入する。
func (w *Worker) Enqueue(job *importjob.Job) {
	w.jobCh <- job
}

func (w *Worker) processJob(ctx context.Context, job *importjob.Job) {
	w.logger.InfoContext(ctx, "import job processing",
		slog.String("job_id", string(job.ID)),
		slog.Int("total_rows", job.TotalRows),
	)
	job.Start()

	// 既存メニュー名を一括取得（外部アクセス1回）
	existingMenus, err := w.menuReader.FindAll(ctx)
	if err != nil {
		w.logger.ErrorContext(ctx, "import job failed to load menus",
			slog.String("job_id", string(job.ID)),
			slog.Any("error", err),
		)
		job.Complete()
		return
	}
	existingNames := make(map[string]bool)
	for _, m := range existingMenus {
		existingNames[m.Name().String()] = true
	}

	// チャンク分割して並列処理
	// existingNames は複数 goroutine から読み書きされるため mutex で保護する。
	rows := job.Rows
	var mu sync.RWMutex
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
			w.processChunk(ctx, job, c, o, existingNames, &mu)
		}(chunk, offset)
	}
	wg.Wait()

	job.Complete()
	w.logger.InfoContext(ctx, "import job completed",
		slog.String("job_id", string(job.ID)),
		slog.String("summary", job.Summary()),
	)
}

func (w *Worker) processChunk(
	ctx context.Context,
	job *importjob.Job,
	chunk []importjob.ImportRow,
	offset int,
	existingNames map[string]bool,
	mu *sync.RWMutex,
) {
	for i, row := range chunk {
		rowNum := offset + i + 1

		// 業務ルール: 同名メニュー重複チェック（インメモリ）
		mu.RLock()
		exists := existingNames[row.MenuName]
		mu.RUnlock()
		if exists {
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

		mu.Lock()
		existingNames[row.MenuName] = true // 同じバッチ内の重複も防ぐ
		mu.Unlock()
		job.RecordSuccess()
	}
}
