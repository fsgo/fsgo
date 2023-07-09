// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/7/2

package fsrpc

import (
	"context"
	"errors"
	"fmt"
)

type AuthHandler struct {
	Method      string
	NewAuthData func(ctx context.Context) *AuthData
	CheckAuth   func(ctx context.Context, ar *AuthData) error
}

func (ah *AuthHandler) getMethod() string {
	if ah.Method != "" {
		return ah.Method
	}
	return "sys_auth"
}

func (ah *AuthHandler) Client(ctx context.Context, rw RequestWriter) (ret error) {
	req := NewRequest(ah.getMethod())
	data := ah.NewAuthData(ctx)
	rr, err := WriteRequestProto(ctx, rw, req, data)
	if err != nil {
		return err
	}
	resp, _, err := rr.Response()
	if err != nil {
		return err
	}
	if resp.GetCode() == ErrCode_Success {
		return nil
	}
	return fmt.Errorf("%w, code=%d, msg=%q", ErrAuthFailed, resp.GetCode(), resp.GetMessage())
}

func (ah *AuthHandler) Server(ctx context.Context, rr RequestReader, rw ResponseWriter) (ret error) {
	req, auth, err := ReadRequestProto(ctx, rr, &AuthData{})
	if err != nil {
		resp := NewResponse(req.GetID(), ErrCode_AuthFailed, "auth failed")
		_ = WriteResponseProto(ctx, rw, resp, nil)
		return err
	}
	err = ah.CheckAuth(ctx, auth)
	if err == nil {
		session := ConnSessionFromCtx(ctx)
		session.LoggedIn.Store(true)
		session.User.Store(auth.UserName)
		return nil
	}
	resp := NewResponse(req.GetID(), ErrCode_AuthFailed, "auth failed")
	_ = WriteResponseProto(ctx, rw, resp, nil)
	return errors.New("auth failed")
}
