// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/7/2

package fsrpc

import (
	"context"
	"fmt"
)

type AuthHandler struct {
	Method      string
	ClientData  func(ctx context.Context) *AuthData
	ServerCheck func(ctx context.Context, ar *AuthData) error
}

func (ah *AuthHandler) RegisterTo(rt RouteRegister) {
	rt.Register(ah.getMethod(), HandlerFunc(ah.Server))
}

func (ah *AuthHandler) getMethod() string {
	if ah.Method != "" {
		return ah.Method
	}
	return "sys_auth"
}

func (ah *AuthHandler) Client(ctx context.Context, rw RequestWriter) (ret error) {
	req := NewRequest(ah.getMethod())
	data := ah.ClientData(ctx)
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
		resp := NewResponse(req.GetID(), ErrCode_AuthFailed, "cannot read auth data: "+err.Error())
		_ = WriteResponseProto(ctx, rw, resp, nil)
		return err
	}
	err = ah.ServerCheck(ctx, auth)
	if err == nil {
		session := ConnSessionFromCtx(ctx)
		session.LoggedIn.Store(true)
		session.User.Store(auth.GetUserName())
		return nil
	}
	resp := NewResponse(req.GetID(), ErrCode_AuthFailed, "check auth: "+err.Error())
	_ = WriteResponseProto(ctx, rw, resp, nil)
	return fmt.Errorf("%w:%w", err, ErrAuthFailed)
}

func (ah *AuthHandler) WithInterceptor(h Handler) Handler {
	it := &Interceptor{
		Name: "auth",
		Before: func(ctx context.Context, rr RequestReader, rw ResponseWriter) (context.Context, RequestReader, ResponseWriter, error) {
			if ah.ServerCheck == nil {
				return ctx, rr, rw, nil
			}
			session := ConnSessionFromCtx(ctx)
			if !session.LoggedIn.Load() {
				req, _ := rr.Request()
				resp := NewResponse(req.GetID(), ErrCode_AuthFailed, "not authed")
				_ = WriteResponseProto(ctx, rw, resp, nil)
				return ctx, rr, rw, ErrAuthFailed
			}
			return ctx, rr, rw, nil
		},
		Handler: h,
	}
	return HandlerFunc(it.Handle)
}
