package handler

import (
	"net/http"

	menucommand "mendo/internal/application/command/menu"
	"mendo/internal/application/query/importstatus"
	importjob "mendo/internal/domain/import"
)

// ImportHandler はインポート関連の HTTP ハンドラ。
type ImportHandler struct {
	importUC *menucommand.ImportMenusUsecase
	statusUC *importstatus.ImportStatusHandler
}

// NewImportHandler は ImportHandler を作成する。
func NewImportHandler(iu *menucommand.ImportMenusUsecase, su *importstatus.ImportStatusHandler) *ImportHandler {
	return &ImportHandler{importUC: iu, statusUC: su}
}

// HandleImportMenus は POST /admin/import/menus のハンドラ。
// CSV を受け取ってジョブを作成し、即座に「受付完了」を返す。
func (h *ImportHandler) HandleImportMenus(w http.ResponseWriter, r *http.Request) error {
	var body struct {
		Items []struct {
			MenuName string `json:"menu_name"`
			Price    int    `json:"price"`
		} `json:"items"`
	}
	if err := readJSON(r, &body); err != nil {
		return err
	}

	rows := make([]importjob.ImportRow, len(body.Items))
	for i, item := range body.Items {
		rows[i] = importjob.ImportRow{MenuName: item.MenuName, Price: item.Price}
	}

	jobID, err := h.importUC.Execute(r.Context(), rows)
	if err != nil {
		return err
	}

	// 202 Accepted（非同期処理を受け付けた）
	WriteSuccess(w, http.StatusAccepted, map[string]string{
		"job_id":  jobID,
		"status":  "queued",
		"message": "インポートを受け付けました。GET /admin/import/{id}/status でプログレスを確認できます",
	})
	return nil
}

// HandleImportStatus は GET /admin/import/{id}/status のハンドラ。
func (h *ImportHandler) HandleImportStatus(w http.ResponseWriter, r *http.Request) error {
	jobID := r.PathValue("id")
	status, err := h.statusUC.Handle(r.Context(), jobID)
	if err != nil {
		return err
	}

	WriteSuccess(w, http.StatusOK, status)
	return nil
}
