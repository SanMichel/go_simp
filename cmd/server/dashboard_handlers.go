package main

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

func (a *App) dashboardPage(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value(ctxUser).(*User)
	activities, _ := a.listActivities(r.Context(), parseFilters(r), 50)
	options, _ := a.listFilterOptions(r.Context())
	a.render(w, "dashboard", map[string]any{"User": u, "Activities": activities, "Options": options, "Filters": parseFilters(r)})
}

func (a *App) dashboardTable(w http.ResponseWriter, r *http.Request) {
	activities, err := a.listActivities(r.Context(), parseFilters(r), 50)
	if err != nil {
		a.handleError(w, r, &AppError{Code: ErrCodeInternal, Message: "Erro ao carregar atividades", HTTPStatus: http.StatusInternalServerError, Err: err})
		return
	}
	options, _ := a.listFilterOptions(r.Context())
	a.render(w, "activities_table", map[string]any{"Activities": activities, "Options": options, "Filters": parseFilters(r)})
}

func (a *App) activityDetails(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	act, items, err := a.activityDetailsData(r.Context(), id)
	if err != nil {
		a.handleError(w, r, &AppError{Code: ErrCodeNotFound, Message: "Atividade não encontrada", HTTPStatus: http.StatusNotFound, Err: err})
		return
	}
	a.render(w, "activity_modal", map[string]any{"Activity": act, "Items": items})
}

func (a *App) printOne(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	a.printActivities(w, r, []int{id})
}

func (a *App) printBulk(w http.ResponseWriter, r *http.Request) {
	var ids []int
	for _, part := range strings.Split(r.URL.Query().Get("ids"), ",") {
		if id, err := strconv.Atoi(strings.TrimSpace(part)); err == nil && id > 0 {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		a.handleError(w, r, &AppError{Code: ErrCodeBadRequest, Message: "IDs inválidos", HTTPStatus: http.StatusBadRequest})
		return
	}
	a.printActivities(w, r, ids)
}

func (a *App) printActivities(w http.ResponseWriter, r *http.Request, ids []int) {
	type Bundle struct {
		Activity Activity
		Items    []ProductVerification
	}
	var bundles []Bundle
	for _, id := range ids {
		act, items, err := a.activityDetailsData(r.Context(), id)
		if err == nil {
			bundles = append(bundles, Bundle{Activity: act, Items: items})
		}
	}
	for _, id := range ids {
		if _, err := a.pg.ExecContext(r.Context(), `UPDATE atividades SET impresso=true WHERE id=$1`, id); err != nil {
			slog.WarnContext(r.Context(), "failed to set impresso for activity",
				"activity_id", id,
				"error", err,
			)
		}
	}
	a.render(w, "print", map[string]any{"Bundles": bundles})
}
