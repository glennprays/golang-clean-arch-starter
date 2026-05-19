package httperror

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/glennprays/golang-clean-arch-starter/internal/apperror"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// validate is shared across BindAndValidate calls; the library is
// goroutine-safe. It's configured to report JSON field names (rather
// than Go struct field names) in validation errors so clients see
// the same identifiers they sent.
var validate = func() *validator.Validate {
	v := validator.New()
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	return v
}()

// BindAndValidate parses the request body into T and validates it
// against its struct tags. On failure it returns a typed *apperror
// that the global error handler will turn into a proper response:
//
//   - parse error              -> 400 BAD_REQUEST
//   - validation failures      -> 400 VALIDATION_FAILED with details[]
//
// Handlers can use it as:
//
//	dto, err := httperror.BindAndValidate[CreateUserDTO](c)
//	if err != nil {
//	    return err
//	}
func BindAndValidate[T any](c *fiber.Ctx) (T, error) {
	var dto T
	if err := c.BodyParser(&dto); err != nil {
		return dto, apperror.BadRequestf("invalid request body: %v", err)
	}
	if err := validate.Struct(&dto); err != nil {
		return dto, fromValidationError(err)
	}
	return dto, nil
}

// fromValidationError converts go-playground/validator failures into an
// *apperror.Error carrying one FieldError per failing field.
func fromValidationError(err error) *apperror.Error {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return apperror.BadRequest(err.Error())
	}
	out := apperror.Validation("request validation failed")
	for _, fe := range ve {
		out.WithDetails(apperror.FieldError{
			Field:   strings.ToLower(stripRootNamespace(fe.Namespace())),
			Rule:    fe.Tag(),
			Message: ruleMessage(fe),
		})
	}
	return out
}

// stripRootNamespace removes the top-level struct name from a
// validator namespace, e.g. "CreateUserDTO.email" -> "email".
func stripRootNamespace(ns string) string {
	if i := strings.Index(ns, "."); i >= 0 {
		return ns[i+1:]
	}
	return ns
}

// ruleMessage produces a short human-readable message for the failed
// rule. Kept deliberately generic — callers wanting custom messages
// can post-process the Details slice.
func ruleMessage(fe validator.FieldError) string {
	if p := fe.Param(); p != "" {
		return fmt.Sprintf("failed on '%s=%s' rule", fe.Tag(), p)
	}
	return fmt.Sprintf("failed on '%s' rule", fe.Tag())
}
