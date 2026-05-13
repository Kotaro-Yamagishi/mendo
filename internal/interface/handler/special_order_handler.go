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

type createSpecialOrderRequest struct {
	OrderID  string `json:"order_id"  validate:"required"`
	MenuName string `json:"menu_name" validate:"required"`
}

type rejectSpecialOrderRequest struct {
	Reason        string `json:"reason"         validate:"required"`
	SuggestedMenu string `json:"suggested_menu" validate:"required"`
}

type resubmitSpecialOrderRequest struct {
	MenuName string `json:"menu_name" validate:"required"`
}

func (h *SpecialOrderHandler) HandleCreate(w http.ResponseWriter, r *http.Request) error {
	var req createSpecialOrderRequest
	if err := readJSON(r, &req); err != nil {
		return err
	}
	if err := validateInput(req); err != nil {
		return err
	}
	id, err := h.createUC.Execute(r.Context(), req.OrderID, req.MenuName)
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
	var req rejectSpecialOrderRequest
	if err := readJSON(r, &req); err != nil {
		return err
	}
	if err := validateInput(req); err != nil {
		return err
	}
	if err := h.rejectUC.Execute(r.Context(), id, req.Reason, req.SuggestedMenu); err != nil {
		return err
	}
	WriteSuccess(w, http.StatusOK, map[string]string{"status": "rejected", "suggested_menu": req.SuggestedMenu})
	return nil
}

func (h *SpecialOrderHandler) HandleResubmit(w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("id")
	var req resubmitSpecialOrderRequest
	if err := readJSON(r, &req); err != nil {
		return err
	}
	if err := validateInput(req); err != nil {
		return err
	}
	if err := h.resubmitUC.Execute(r.Context(), id, req.MenuName); err != nil {
		return err
	}
	WriteSuccess(w, http.StatusOK, map[string]string{"status": "pending"})
	return nil
}
