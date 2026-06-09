package main

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func (a *App) adminPage(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value(ctxUser).(*User)
	users, err := a.listUsers(r.Context())
	if err != nil {
		a.handleError(w, r, &AppError{Code: ErrCodeInternal, Message: "Erro ao carregar usuários", HTTPStatus: http.StatusInternalServerError, Err: err})
		return
	}
	a.render(w, "admin", map[string]any{"User": u, "Users": users})
}

func (a *App) adminUsersSection(w http.ResponseWriter, r *http.Request) {
	users, _ := a.listUsers(r.Context())
	a.render(w, "users_section", map[string]any{"Users": users})
}

func (a *App) adminCreateUser(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		a.handleError(w, r, &AppError{Code: ErrCodeBadRequest, Message: "Erro ao processar formulário", HTTPStatus: http.StatusBadRequest, Err: err})
		return
	}
	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")
	role := r.FormValue("role")
	if username == "" || len(password) < 8 || !validRole(role) {
		a.handleError(w, r, &AppError{Code: ErrCodeValidation, Message: "Dados inválidos.", HTTPStatus: http.StatusBadRequest})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		a.handleError(w, r, &AppError{Code: ErrCodeInternal, Message: "Erro interno do servidor", HTTPStatus: http.StatusInternalServerError})
		return
	}
	_, err = a.pg.ExecContext(r.Context(), `INSERT INTO users (username, password, role) VALUES ($1,$2,$3)`, username, string(hash), role)
	users, _ := a.listUsers(r.Context())
	if err != nil {
		slog.ErrorContext(r.Context(), "error creating user", "error", err)
		a.render(w, "users_section", map[string]any{"Users": users, "Message": "Erro interno do servidor", "Error": true})
		return
	}
	a.render(w, "users_section", map[string]any{"Users": users, "Message": "Usuário criado com sucesso."})
}

func (a *App) adminEditUserRow(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	u, err := a.findUserByID(r.Context(), id)
	if err != nil {
		a.handleError(w, r, &AppError{Code: ErrCodeNotFound, Message: "Usuário não encontrado", HTTPStatus: http.StatusNotFound, Err: err})
		return
	}
	a.render(w, "user_edit_row", map[string]any{"RowUser": u})
}

func (a *App) adminUserRow(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	u, err := a.findUserByID(r.Context(), id)
	if err != nil {
		a.handleError(w, r, &AppError{Code: ErrCodeNotFound, Message: "Usuário não encontrado", HTTPStatus: http.StatusNotFound, Err: err})
		return
	}
	a.render(w, "user_row", map[string]any{"RowUser": u})
}

func (a *App) updateUserAdmin(ctx context.Context, currentUserID, targetID int, role, password string) (*UserRow, error) {
	if targetID == currentUserID {
		return nil, &AppError{Code: ErrCodeBadRequest, Message: "Não é possível editar o próprio usuário", HTTPStatus: http.StatusBadRequest}
	}
	if !validRole(role) {
		return nil, &AppError{Code: ErrCodeBadRequest, Message: "Role inválido", HTTPStatus: http.StatusBadRequest}
	}
	target, err := a.findUserByID(ctx, targetID)
	if err != nil {
		return nil, &AppError{Code: ErrCodeNotFound, Message: "Usuário não encontrado", HTTPStatus: http.StatusNotFound, Err: err}
	}
	currentUser, err := a.findUserByID(ctx, currentUserID)
	if err != nil {
		return nil, &AppError{Code: ErrCodeNotFound, Message: "Usuário não encontrado", HTTPStatus: http.StatusNotFound, Err: err}
	}
	if target.Role == "sysadmin" && currentUser.Role != "sysadmin" {
		return nil, &AppError{Code: ErrCodeForbidden, Message: "Sem permissão", HTTPStatus: http.StatusForbidden}
	}
	if password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, &AppError{Code: ErrCodeInternal, Message: "Erro interno do servidor", HTTPStatus: http.StatusInternalServerError}
		}
		if _, err := a.pg.ExecContext(ctx, `UPDATE users SET role=$1, password=$2, last_token_at=now() WHERE id=$3`, role, string(hash), targetID); err != nil {
			slog.ErrorContext(ctx, "error updating user", "user_id", targetID, "error", err)
		}
	} else {
		if _, err := a.pg.ExecContext(ctx, `UPDATE users SET role=$1, last_token_at=now() WHERE id=$2`, role, targetID); err != nil {
			slog.ErrorContext(ctx, "error updating user", "user_id", targetID, "error", err)
		}
	}
	found, err := a.findUserByID(ctx, targetID)
	if err != nil {
		return nil, err
	}
	return &UserRow{ID: found.ID, Username: found.Username, Role: found.Role, LastTokenAt: found.LastTokenAt}, nil
}

func (a *App) adminUpdateUser(w http.ResponseWriter, r *http.Request) {
	currentUser := r.Context().Value(ctxUser).(*User)
	id, _ := strconv.Atoi(r.PathValue("id"))
	if err := r.ParseForm(); err != nil {
		a.handleError(w, r, &AppError{Code: ErrCodeBadRequest, Message: "Erro ao processar formulário", HTTPStatus: http.StatusBadRequest, Err: err})
		return
	}
	role := r.FormValue("role")
	password := r.FormValue("password")
	u, err := a.updateUserAdmin(r.Context(), currentUser.ID, id, role, password)
	if err != nil {
		a.handleError(w, r, err)
		return
	}
	a.render(w, "user_row", map[string]any{"RowUser": u})
}
