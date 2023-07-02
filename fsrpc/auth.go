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
	Method         string
	NewAuthRequest func(ctx context.Context) *AuthRequest
	CheckAuth      func(ctx context.Context, ar *AuthRequest) error
}

func (ah *AuthHandler) getMethod() string {
	if ah.Method != "" {
		return ah.Method
	}
	return "sys_auth"
}

func (ah *AuthHandler) Send(ctx context.Context, rw RequestProtoWriter) (ret error) {
	req := NewRequest(ah.getMethod())
	data := ah.NewAuthRequest(ctx)
	rr, err := rw.Write(ctx, req, data)
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

func (ah *AuthHandler) Receiver(ctx context.Context, rr RequestReader, rw ResponseWriter) (ret error) {
	req, auth, err := ReadProtoRequest(ctx, rr, &AuthRequest{})
	if err != nil {
		resp := NewResponse(req.GetID(), ErrCode_ReqNoAuth, "auth failed")
		rw.Write(ctx, resp, nil)
		return err
	}
	err = ah.CheckAuth(ctx, auth)
	if err == nil {
		session := ServerConnSessionFromCtx(ctx)
		session.LoggedIn.Store(true)
		session.User.Store(auth.UserName)
		return nil
	}
	resp := NewResponse(req.GetID(), ErrCode_ReqNoAuth, "auth failed")
	rw.Write(ctx, resp, nil)
	return errors.New("auth failed")
}
