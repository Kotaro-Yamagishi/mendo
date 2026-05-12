package handler

import (
	"net/http"

	socommand "mendo/internal/application/command/specialorder"
)

type SpecialOrderHandler struct {
	createUC   *socommand.CreateSpecialOrderUsecase
	approveUC  *socommand.ApproveSpecialOrderUsecase
	rejectUC   *socommand.RejectSpecialOrderUsecase
	resubmitUC *socommand.ResubmitSpecialOrderUsecase
}

func NewSpecialOrderHandler(
	c *socommand.CreateSpecialOrderUsecase,
	a *socommand.ApproveSpecialOrderUsecase,
	r *socommand.RejectSpecialOrderUsecase,
	rs *socommand.ResubmitSpecialOrderUsecase,
) *SpecialOrderHandler {
	return &SpecialOrderHandler{createUC: c, approveUC: a, rejectUC: r, resubmitUC: rs}
}

func (h *SpecialOrderHandler) HandleCreate(w http.ResponseWriter, r *http.Request) error {
	var body struct {
		OrderID  string `json:"order_id"`
		MenuName string `json:"menu_name"`
	}
	if err := readJSON(r, &body); err != nil {
		return err
	}
	id, err := h.createUC.Execute(r.Context(), body.OrderID, body.MenuName)
	if err != nil {
		return err
	}
	WriteSuccess(w, http.StatusCreated, map[string]string{"special_order_id": id, "status": "pending"})
	return nil
}

func (h *SpecialOrderHandler) HandleApprove(w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("id")
	if err := h.approveUC.Execute(r.Context(), id); err != nil {
		return err
	}
	WriteSuccess(w, http.StatusOK, map[string]string{"status": "approved_and_cooking"})
	return nil
}

func (h *SpecialOrderHandler) HandleReject(w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("id")
	var body struct {
		Reason        string `json:"reason"`
		SuggestedMenu string `json:"suggested_menu"`
	}
	if err := readJSON(r, &body); err != nil {
		return err
	}
	if err := h.rejectUC.Execute(r.Context(), id, body.Reason, body.SuggestedMenu); err != nil {
		return err
	}
	WriteSuccess(w, http.StatusOK, map[string]string{"status": "rejected", "suggested_menu": body.SuggestedMenu})
	return nil
}

func (h *SpecialOrderHandler) HandleResubmit(w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("id")
	var body struct {
		MenuName string `json:"menu_name"`
	}
	if err := readJSON(r, &body); err != nil {
		return err
	}
	if err := h.resubmitUC.Execute(r.Context(), id, body.MenuName); err != nil {
		return err
	}
	WriteSuccess(w, http.StatusOK, map[string]string{"status": "pending"})
	return nil
}
