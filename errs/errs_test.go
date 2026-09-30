package errs_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/lpphub/gost/errs"
)

func TestNewStatus(t *testing.T) {
	cases := []struct {
		name   string
		err    *errs.Error
		status int
	}{
		{"New 默认", errs.New(1, "m"), 200},
		{"NewWithStatus(0) 回落到默认", errs.NewWithStatus(1, "m", 0), 200},
		{"NewWithStatus(404)", errs.NewWithStatus(1, "m", 404), 404},
		{"ErrUnclassified", errs.ErrUnclassified, 500},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.err.Status(); got != c.status {
				t.Fatalf("Status() = %d, want %d", got, c.status)
			}
		})
	}
}

func TestErrUnclassified(t *testing.T) {
	if got, want := errs.ErrUnclassified.Code(), -1; got != want {
		t.Fatalf("Code() = %d, want %d", got, want)
	}
	if got, want := errs.ErrUnclassified.Message(), "internal error"; got != want {
		t.Fatalf("Message() = %q, want %q", got, want)
	}
	if errs.ErrUnclassified.Cause() != nil {
		t.Fatalf("Cause() = %v, want nil", errs.ErrUnclassified.Cause())
	}
}

func TestErrorString(t *testing.T) {
	sqlErr := errors.New("sql: no rows in result set")

	if got, want := errs.New(1001, "用户不存在").Error(), "用户不存在"; got != want {
		t.Fatalf("无 cause: got %q, want %q", got, want)
	}
	if got, want := errs.New(1001, "用户不存在").Wrap(sqlErr).Error(),
		"用户不存在: sql: no rows in result set"; got != want {
		t.Fatalf("有 cause: got %q, want %q", got, want)
	}

	e := errs.ErrUnclassified.Wrapf("panic: boom")
	if got, want := e.Message(), "internal error"; got != want {
		t.Fatalf("Message() 泄露了内部信息: %q", got)
	}
	if !strings.Contains(e.Error(), "panic: boom") {
		t.Fatalf("Error() 丢了原因: %q", e.Error())
	}
}

func TestWrapImmutable(t *testing.T) {
	base := errs.New(1001, "用户不存在")

	wrapped := base.Wrap(errors.New("boom"))
	wrappedf := base.Wrapf("boom %d", 1)

	if wrapped.Cause() == nil || wrappedf.Cause() == nil {
		t.Fatal("Wrap/Wrapf 应产出带 cause 的副本")
	}
	if base.Cause() != nil {
		t.Fatalf("Wrap 改到了原对象: Cause() = %v", base.Cause())
	}
	if got, want := base.Error(), "用户不存在"; got != want {
		t.Fatalf("Wrap 改到了原对象: Error() = %q", got)
	}
	if base.Wrap(nil) != base {
		t.Fatal("Wrap(nil) 应原样返回")
	}
}

func TestWrapf(t *testing.T) {
	err := errs.New(1001, "用户不存在").Wrapf("uid=%d: %w", 42, errors.New("boom"))
	if got, want := err.Error(), "用户不存在: uid=42: boom"; got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}

func TestWith(t *testing.T) {
	base := errs.NewWithStatus(1003, "订单已锁定", http.StatusConflict)
	sqlErr := errors.New("sql: no rows in result set")

	got := base.With("订单 %s 已锁定", "o1")
	if want := "订单 o1 已锁定"; got.Message() != want {
		t.Fatalf("Message() = %q, want %q", got.Message(), want)
	}
	if got.Code() != base.Code() || got.Status() != base.Status() {
		t.Fatalf("With 改了 code/status: %d/%d", got.Code(), got.Status())
	}
	if base.Message() != "订单已锁定" {
		t.Fatalf("With 改到了原对象: %q", base.Message())
	}
	if !errors.Is(got, base) {
		t.Fatal("With 派生后仍应按 code 命中")
	}

	both := base.With("订单 %s 已锁定", "o1").Wrap(sqlErr)
	if both.Message() != "订单 o1 已锁定" || both.Cause() != sqlErr {
		t.Fatalf("With + Wrap 组合不对: msg=%q cause=%v", both.Message(), both.Cause())
	}
	if got, want := both.Error(), "订单 o1 已锁定: sql: no rows in result set"; got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}

func TestUnwrap(t *testing.T) {
	sqlErr := errors.New("sql: no rows in result set")
	e := errs.New(1001, "用户不存在").Wrap(sqlErr)

	if !errors.Is(e, sqlErr) {
		t.Fatal("errors.Is 没有穿透到 cause")
	}
	if !errors.Is(fmt.Errorf("svc: %w", e), sqlErr) {
		t.Fatal("多层包装后仍应穿透到 cause")
	}
}

func TestIs(t *testing.T) {
	base := errs.New(1001, "用户不存在")

	cases := []struct {
		name   string
		err    error
		target error
		want   bool
	}{
		{"同码不同实例", base, errs.New(1001, "换个文案"), true},
		{"不同码", base, errs.New(1002, "余额不足"), false},
		{"目标不是 *Error", base, errors.New("x"), false},
		{"err 被 fmt.Errorf 包过", fmt.Errorf("svc: %w", base), base, true},
		{"err 带 cause", base.Wrap(errors.New("boom")), base, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := errors.Is(c.err, c.target); got != c.want {
				t.Fatalf("errors.Is() = %v, want %v", got, c.want)
			}
		})
	}
}

func TestAs(t *testing.T) {
	base := errs.New(1001, "用户不存在")

	if got, ok := errs.As(fmt.Errorf("svc: %w", base)); !ok || got != base {
		t.Fatalf("As() = (%v, %v), want (%v, true)", got, ok, base)
	}
	if _, ok := errs.As(errors.New("x")); ok {
		t.Fatal("As 不该命中普通错误")
	}
	if _, ok := errs.As(nil); ok {
		t.Fatal("As(nil) 不该命中")
	}
}

func TestIsCode(t *testing.T) {
	if !errs.IsCode(fmt.Errorf("svc: %w", errs.New(1001, "用户不存在")), 1001) {
		t.Fatal("IsCode 应穿透包装")
	}
	if errs.IsCode(errs.New(1001, "用户不存在"), 1002) {
		t.Fatal("不同码不该命中")
	}
	if errs.IsCode(errors.New("x"), 1001) {
		t.Fatal("普通错误不该命中")
	}
}

func TestNormalize(t *testing.T) {
	base := errs.New(1001, "用户不存在")
	sqlErr := errors.New("dial tcp: connection refused")

	if got := errs.Normalize(nil); got != nil {
		t.Fatalf("Normalize(nil) = %v, want nil", got)
	}
	if got := errs.Normalize(base); got != base {
		t.Fatalf("*Error 应原样返回, got %v", got)
	}
	if got := errs.Normalize(fmt.Errorf("svc: %w", base)); got != base {
		t.Fatalf("包过的 *Error 应取出内层, got %v", got)
	}

	got := errs.Normalize(sqlErr)
	if got.Code() != -1 || got.Status() != 500 || got.Message() != "internal error" {
		t.Fatalf("兜底错误不对: code=%d status=%d msg=%q", got.Code(), got.Status(), got.Message())
	}
	if got.Cause() != sqlErr {
		t.Fatalf("原因没保留: Cause() = %v", got.Cause())
	}
	if !errors.Is(got, errs.ErrUnclassified) {
		t.Fatal("兜底结果应 Is ErrUnclassified")
	}
	if errs.ErrUnclassified.Cause() != nil {
		t.Fatalf("Normalize 污染了 ErrUnclassified: %v", errs.ErrUnclassified.Cause())
	}
}

func TestNormalizeNeverNil(t *testing.T) {
	errsToCheck := []error{
		errors.New("boom"),
		fmt.Errorf("a: %w", errors.New("b")),
		context.Canceled,
		context.DeadlineExceeded,
	}
	for _, err := range errsToCheck {
		if got := errs.Normalize(err); got == nil {
			t.Fatalf("Normalize(%v) = nil, 契约要求非 nil", err)
		}
	}
}
