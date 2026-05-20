package apperror_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/glennprays/golang-clean-arch-starter/internal/apperror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestError_NilSafe(t *testing.T) {
	var e *apperror.Error // typed nil

	assert.Equal(t, "", e.Error())
	assert.Nil(t, e.Unwrap())
	assert.Equal(t, http.StatusInternalServerError, e.HTTPStatus())
	assert.Nil(t, e.Wrap(errors.New("anything")))
	assert.Nil(t, e.WithDetails(apperror.FieldError{Field: "x"}))
}

func TestIs_MatchesByKind(t *testing.T) {
	cases := []struct {
		name   string
		err    *apperror.Error
		target *apperror.Error
		want   bool
	}{
		{"same kind, constructed", apperror.NotFoundf("user %d", 42), apperror.ErrNotFound, true},
		{"different kind", apperror.BadRequest("bad"), apperror.ErrNotFound, false},
		{"validation matches sentinel", apperror.Validation("x"), apperror.ErrValidation, true},
		{"internal matches sentinel", apperror.Internalf("db %s", "down"), apperror.ErrInternal, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, errors.Is(tc.err, tc.target))
		})
	}
}

func TestIs_WorksThroughWrap(t *testing.T) {
	cause := errors.New("pgx: no rows")
	err := apperror.NotFoundf("user %d", 42).Wrap(cause)

	require.True(t, errors.Is(err, apperror.ErrNotFound), "should match by Kind through wrap")
	require.True(t, errors.Is(err, cause), "should reach cause via Unwrap")
}

func TestUnwrap_ReturnsCause(t *testing.T) {
	cause := errors.New("boom")
	err := apperror.Internal("db failed").Wrap(cause)

	assert.Same(t, cause, errors.Unwrap(err))
}

func TestUnwrap_NoCause(t *testing.T) {
	err := apperror.NotFound("nope")
	assert.Nil(t, errors.Unwrap(err))
}

func TestHTTPStatus_AllKinds(t *testing.T) {
	cases := map[apperror.Kind]int{
		apperror.KindBadRequest:   http.StatusBadRequest,
		apperror.KindValidation:   http.StatusBadRequest,
		apperror.KindUnauthorized: http.StatusUnauthorized,
		apperror.KindForbidden:    http.StatusForbidden,
		apperror.KindNotFound:     http.StatusNotFound,
		apperror.KindConflict:     http.StatusConflict,
		apperror.KindInternal:     http.StatusInternalServerError,
	}
	for kind, want := range cases {
		t.Run(string(kind), func(t *testing.T) {
			e := &apperror.Error{Kind: kind, Message: "x"}
			assert.Equal(t, want, e.HTTPStatus())
		})
	}
}

func TestHTTPStatus_UnknownKind(t *testing.T) {
	e := &apperror.Error{Kind: "TOTALLY_MADE_UP", Message: "x"}
	assert.Equal(t, http.StatusInternalServerError, e.HTTPStatus())
}

func TestConstructors_SetKindAndMessage(t *testing.T) {
	cases := []struct {
		name string
		got  *apperror.Error
		kind apperror.Kind
		msg  string
	}{
		{"BadRequest", apperror.BadRequest("bad"), apperror.KindBadRequest, "bad"},
		{"NotFound", apperror.NotFound("nope"), apperror.KindNotFound, "nope"},
		{"NotFoundf", apperror.NotFoundf("user %d", 42), apperror.KindNotFound, "user 42"},
		{"Conflict", apperror.Conflict("dup"), apperror.KindConflict, "dup"},
		{"Internal", apperror.Internal("db"), apperror.KindInternal, "db"},
		{"Internalf", apperror.Internalf("%s failed", "db"), apperror.KindInternal, "db failed"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.kind, tc.got.Kind)
			assert.Equal(t, tc.msg, tc.got.Message)
		})
	}
}

func TestWithDetails_Appends(t *testing.T) {
	e := apperror.Validation("invalid").
		WithDetails(apperror.FieldError{Field: "email", Rule: "email"}).
		WithDetails(apperror.FieldError{Field: "age", Rule: "min"})

	require.Len(t, e.Details, 2)
	assert.Equal(t, "email", e.Details[0].Field)
	assert.Equal(t, "age", e.Details[1].Field)
}

func TestWrap_AttachesCause(t *testing.T) {
	cause := errors.New("downstream")
	e := apperror.Internal("db failed").Wrap(cause)

	assert.Same(t, cause, e.Cause)
}

func TestError_String_IncludesCauseWhenPresent(t *testing.T) {
	withCause := apperror.Internal("db failed").Wrap(errors.New("conn refused"))
	withoutCause := apperror.NotFound("user 42")

	assert.Contains(t, withCause.Error(), "INTERNAL")
	assert.Contains(t, withCause.Error(), "db failed")
	assert.Contains(t, withCause.Error(), "conn refused")

	assert.Contains(t, withoutCause.Error(), "NOT_FOUND")
	assert.Contains(t, withoutCause.Error(), "user 42")
	assert.NotContains(t, withoutCause.Error(), "<nil>")
}

func TestAsAppError_ExtractsFromChain(t *testing.T) {
	inner := apperror.NotFoundf("user %d", 7)
	// fmt.Errorf %w produces a stdlib wrapper around inner.
	wrapped := fmt.Errorf("usecase: %w", inner)

	got := apperror.AsAppError(wrapped)
	require.NotNil(t, got)
	assert.Equal(t, apperror.KindNotFound, got.Kind)
}

func TestAsAppError_NotInChain(t *testing.T) {
	plain := errors.New("plain")
	assert.Nil(t, apperror.AsAppError(plain))
}

func TestIs_NilTarget(t *testing.T) {
	err := apperror.NotFound("x")
	assert.False(t, errors.Is(err, nil))
}

func TestIs_NonAppErrorTarget(t *testing.T) {
	err := apperror.NotFound("x")
	stdErr := errors.New("other")
	assert.False(t, errors.Is(err, stdErr))
}
