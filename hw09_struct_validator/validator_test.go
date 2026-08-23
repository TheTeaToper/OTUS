package hw09structvalidator

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
)

type UserRole string

// Test the function on different structures and other types.
type (
	User struct {
		ID     string `json:"id" validate:"len:36"`
		Name   string
		Age    int             `validate:"min:18|max:50"`
		Email  string          `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole        `validate:"in:admin,stuff"`
		Phones []string        `validate:"len:11"`
		meta   json.RawMessage //nolint:unused
	}

	App struct {
		Version string `validate:"len:5"`
	}

	Token struct {
		Header    []byte
		Payload   []byte
		Signature []byte
	}

	Response struct {
		Code int    `validate:"in:200,404,500"`
		Body string `json:"omitempty"`
	}
)

//nolint:gocognit,funlen
func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		in          interface{}
		wantErr     bool
		expectedErr error                         // Для проверки конкретных одиночных ошибок (например, ErrNotAStruct)
		checkFunc   func(t *testing.T, err error) // Для глубокой проверки ValidationErrors
	}{
		{
			name: "valid user and pointer to struct",
			in: &User{
				ID:     "12345678-1234-1234-1234-123456789012", // len: 36
				Name:   "Ivan",                                 // no tags
				Age:    25,                                     // min:18, max:50
				Email:  "test@example.com",                     // regexp match
				Role:   "admin",                                // in:admin,stuff
				Phones: []string{"79991112233", "79991112234"}, // elements len: 11
			},
			wantErr: false,
		},
		{
			name:        "not a struct error",
			in:          "just a string",
			wantErr:     true,
			expectedErr: ErrNotAStruct,
		},
		{
			name: "multiple validation errors accumulation",
			in: User{
				ID:     "short-id",                     // invalid len (8 instead of 36)
				Age:    15,                             // below min 18
				Email:  "invalid-email",                // regex mismatch
				Role:   "guest",                        // not in set
				Phones: []string{"123", "79991112233"}, // first phone is short (len 3)
			},
			wantErr: true,
			checkFunc: func(t *testing.T, err error) {
				t.Helper()
				var vErrs ValidationErrors
				if !errors.As(err, &vErrs) {
					t.Fatalf("expected error to be of type ValidationErrors, got %T", err)
				} // Должно накопиться ровно 5 ошибок валидации
				if len(vErrs) != 5 {
					t.Errorf("expected 5 validation errors, got %d: %v", len(vErrs), vErrs)
				}

				// Проверяем наличие специфичных ошибок через errors.Is внутри слайса
				errMap := make(map[string]error)
				for _, vErr := range vErrs {
					errMap[vErr.Field] = vErr.Err
				}

				if !errors.Is(errMap["ID"], ErrInvalidLen) {
					t.Errorf("expected ErrInvalidLen for ID, got %v", errMap["ID"])
				}
				if !errors.Is(errMap["Age"], ErrMinBound) {
					t.Errorf("expected ErrMinBound for Age, got %v", errMap["Age"])
				}
				if !errors.Is(errMap["Email"], ErrRegexpMatch) {
					t.Errorf("expected ErrRegexpMatch for Email, got %v", errMap["Email"])
				}
				if !errors.Is(errMap["Role"], ErrNotInSet) {
					t.Errorf("expected ErrNotInSet for Role, got %v", errMap["Role"])
				}
				if !errors.Is(errMap["Phones[0]"], ErrInvalidLen) {
					t.Errorf("expected ErrInvalidLen for Phones[0], got %v", errMap["Phones[0]"])
				}
			},
		},
		{
			name: "struct with unsupported tag syntax (program error)",
			in: struct {
				BadField int `validate:"min:not-a-number"`
			}{BadField: 5},
			wantErr: true,
			checkFunc: func(t *testing.T, err error) {
				t.Helper()
				var pErr ProgramError
				if !errors.As(err, &pErr) {
					t.Errorf("expected ProgramError due to bad tag format, got %T", err)
				}
			},
		},
		{
			name: "struct with unhandled field type (program error)",
			in: struct {
				BoolField bool `validate:"min:1"`
			}{BoolField: true},
			wantErr: true,
			checkFunc: func(t *testing.T, err error) {
				t.Helper()
				var pErr ProgramError
				if !errors.As(err, &pErr) {
					t.Errorf("expected ProgramError due to unsupported type, got %T", err)
				}
			},
		},
		{
			name: "response structure valid code",
			in: Response{
				Code: 404,
				Body: "Not Found",
			},
			wantErr: false,
		},
		{
			name: "response structure invalid code",
			in: Response{
				Code: 403,
			},
			wantErr: true,
			checkFunc: func(t *testing.T, err error) {
				t.Helper()
				var vErrs ValidationErrors
				if errors.As(err, &vErrs) {
					if !errors.Is(vErrs[0].Err, ErrNotInSet) {
						t.Errorf("expected ErrNotInSet for Code, got %v", vErrs[0].Err)
					}
				} else {
					t.Errorf("expected ValidationErrors, got %T", err)
				}
			},
		},
		{
			name: "token structure without validation tags",
			in: Token{
				Header: []byte("abc"),
			},
			wantErr: false,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			tt := tt
			t.Parallel()

			err := Validate(tt.in)

			if !tt.wantErr {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}

			if err == nil {
				t.Fatal("expected error, but got nil")
			}

			if tt.expectedErr != nil && !errors.Is(err, tt.expectedErr) {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, err)
			}
			_ = tt
		})
	}
}
