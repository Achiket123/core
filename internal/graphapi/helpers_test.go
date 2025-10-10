package graphapi

import (
	"io"
	"strings"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/samber/lo"
	"github.com/vektah/gqlparser/v2/ast"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/internal/ent/generated"
	pkgobjects "github.com/theopenlane/core/pkg/objects"
)

func TestStripOperation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Create operation",
			input:    "createUser",
			expected: "user",
		},
		{
			name:     "Update operation",
			input:    "updateUser",
			expected: "user",
		},
		{
			name:     "Delete operation",
			input:    "deleteUser",
			expected: "user",
		},
		{
			name:     "Get operation",
			input:    "getUser",
			expected: "user",
		},
		{
			name:     "No operation",
			input:    "User",
			expected: "user",
		},
		{
			name:     "Non-matching prefix",
			input:    "fetchUser",
			expected: "fetch_user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripOperation(tt.input)
			assert.Check(t, is.Equal(tt.expected, result))
		})
	}
}

func TestRetrieveObjectDetails(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		fieldName   string
		key         string
		arguments   ast.ArgumentList
		expected    *pkgobjects.File
		expectedErr error
	}{
		{
			name:      "Matching upload argument",
			fieldName: "createUser",
			key:       "file",
			arguments: ast.ArgumentList{
				&ast.Argument{
					Name: "file",
					Value: &ast.Value{
						ExpectedType: &ast.Type{
							NamedType: "Upload",
						},
					},
				},
			},
			expected: &pkgobjects.File{
				CorrelatedObjectType: "user",
				FileMetadata: pkgobjects.FileMetadata{
					Key: "file",
				},
			},
			expectedErr: nil,
		},
		{
			name:      "Non-matching upload argument",
			fieldName: "createUser",
			key:       "image",
			arguments: ast.ArgumentList{
				&ast.Argument{
					Name: "file",
					Value: &ast.Value{
						ExpectedType: &ast.Type{
							NamedType: "Upload",
						},
					},
				},
			},
			expected:    &pkgobjects.File{},
			expectedErr: ErrUnableToDetermineObjectType,
		},
		{
			name:        "No upload argument",
			fieldName:   "createUser",
			key:         "file",
			arguments:   ast.ArgumentList{},
			expected:    &pkgobjects.File{},
			expectedErr: ErrUnableToDetermineObjectType,
		},
		{
			name:      "Non-upload argument",
			fieldName: "createUser",
			key:       "file",
			arguments: ast.ArgumentList{
				&ast.Argument{
					Name: "file",
					Value: &ast.Value{
						ExpectedType: &ast.Type{
							NamedType: "String",
						},
					},
				},
			},
			expected:    &pkgobjects.File{},
			expectedErr: ErrUnableToDetermineObjectType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rctx := &graphql.FieldContext{
				Field: graphql.CollectedField{
					Field: &ast.Field{
						Name:      tt.fieldName,
						Arguments: tt.arguments,
					},
				},
			}

			upload := &pkgobjects.File{
				OriginalName: "meow.txt",
			}

			result, err := retrieveObjectDetails(rctx, tt.key, upload)
			if tt.expectedErr != nil {

				return
			}

			assert.NilError(t, err)
			assert.Check(t, is.Equal(tt.expected.CorrelatedObjectType, result.CorrelatedObjectType))
			assert.Check(t, is.Equal(tt.expected.FileMetadata.Key, result.FileMetadata.Key))
		})
	}
}
func TestGetOrgOwnerFromInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       any
		expected    *string
		expectedErr error
	}{
		{
			name:        "Nil input",
			input:       nil,
			expected:    nil,
			expectedErr: nil,
		},
		{
			name: "Valid input with owner ID",
			input: generated.CreateProcedureInput{
				Name:    "Test Procedure",
				OwnerID: lo.ToPtr("owner123"),
			},
			expected:    lo.ToPtr("owner123"),
			expectedErr: nil,
		},
		{
			name:  "Valid input without owner ID",
			input: generated.CreateRiskInput{
				// No OwnerID field set
			},
			expected:    nil,
			expectedErr: nil,
		},
		{
			name: "Invalid input type will return nil",
			input: struct {
				Name string `json:"name"`
			}{
				Name: "test",
			},
			expected:    nil,
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getOrgOwnerFromInput(&tt.input)
			if tt.expectedErr != nil {

				assert.Check(t, is.Nil(result))
				return
			}

			assert.NilError(t, err)
			assert.Check(t, is.DeepEqual(tt.expected, result))
		})
	}
}
func TestGetBulkUploadOwnerInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       []*generated.CreateProcedureInput
		expected    *string
		expectedErr error
	}{
		{
			name:        "Nil input, nothing to do",
			input:       nil,
			expected:    nil,
			expectedErr: nil,
		},
		{
			name: "Valid input with consistent owner IDs",
			input: []*generated.CreateProcedureInput{
				{
					Name:    "Test Procedure 1",
					OwnerID: lo.ToPtr("owner123"),
				},
				{
					Name:    "Test Procedure 2",
					OwnerID: lo.ToPtr("owner123"),
				},
			},
			expected:    lo.ToPtr("owner123"),
			expectedErr: nil,
		},
		{
			name: "Valid input with inconsistent owner IDs",
			input: []*generated.CreateProcedureInput{
				{
					Name:    "Test Procedure 1",
					OwnerID: lo.ToPtr("owner123"),
				},
				{
					Name:    "Test Procedure 2",
					OwnerID: lo.ToPtr("owner456"),
				},
			},
			expected:    nil,
			expectedErr: ErrNoOrganizationID,
		},
		{
			name: "Valid input with missing owner ID",
			input: []*generated.CreateProcedureInput{
				{
					Name: "Test Procedure 1",
				},
			},
			expected:    nil,
			expectedErr: ErrNoOrganizationID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getBulkUploadOwnerInput(tt.input)
			if tt.expectedErr != nil {

				assert.Check(t, is.Nil(result))
				return
			}

			assert.NilError(t, err)
			assert.Check(t, is.DeepEqual(tt.expected, result))
		})
	}
}

func TestFileBuffering(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		fileSize           int
		expectBuffered     bool
		expectOriginalKept bool
	}{
		{
			name:           "Small file gets buffered",
			fileSize:       1024,
			expectBuffered: true,
		},
		{
			name:               "Large file exceeds limit",
			fileSize:           pkgobjects.MaxInMemorySize + 1,
			expectOriginalKept: true,
		},
		{
			name:           "File at max size gets buffered",
			fileSize:       pkgobjects.MaxInMemorySize,
			expectBuffered: true,
		},
		{
			name:           "Empty file gets buffered",
			fileSize:       0,
			expectBuffered: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := strings.Repeat("a", tt.fileSize)
			reader := strings.NewReader(data)

			file := pkgobjects.File{
				RawFile: reader,
			}

			// Simulate the buffering logic from injectFileUploader
			if file.RawFile != nil {
				buffered, err := pkgobjects.NewBufferedReaderFromReader(file.RawFile)
				if err == nil {
					file.RawFile = buffered
				}
			}

			if tt.expectBuffered {
				_, ok := file.RawFile.(*pkgobjects.BufferedReader)
				assert.Check(t, ok, "expected BufferedReader but got different type")

				readData, err := io.ReadAll(file.RawFile)
				assert.NilError(t, err)
				assert.Check(t, is.Equal(tt.fileSize, len(readData)))
			}

			if tt.expectOriginalKept {
				_, ok := file.RawFile.(*pkgobjects.BufferedReader)
				assert.Check(t, !ok, "expected original reader but got BufferedReader")
			}
		})
	}
}

func TestFileBufferingWithReadSeeker(t *testing.T) {
	t.Parallel()

	data := []byte("test data")
	bufferedReader := pkgobjects.NewBufferedReader(data)

	file := pkgobjects.File{
		RawFile:      bufferedReader,
		OriginalName: "test.txt",
	}

	// Simulate the buffering logic - should keep the existing BufferedReader
	if file.RawFile != nil {
		newBuffered, err := pkgobjects.NewBufferedReaderFromReader(file.RawFile)
		if err == nil {
			file.RawFile = newBuffered
		} else if err != pkgobjects.ErrFileSizeExceedsLimit {
			t.Errorf("unexpected error: %v", err)
		}
	}

	result, ok := file.RawFile.(*pkgobjects.BufferedReader)
	assert.Check(t, ok, "expected BufferedReader")
	assert.Check(t, is.Equal(int64(len(data)), result.Size()))
}

func TestFileBufferingNilReader(t *testing.T) {
	t.Parallel()

	file := pkgobjects.File{
		RawFile:      nil,
		OriginalName: "test.txt",
	}

	// Simulate the buffering logic with nil reader
	if file.RawFile != nil {
		buffered, err := pkgobjects.NewBufferedReaderFromReader(file.RawFile)
		if err == nil {
			file.RawFile = buffered
		} else if err != pkgobjects.ErrFileSizeExceedsLimit {
			t.Errorf("unexpected error: %v", err)
		}
	}

	assert.Check(t, file.RawFile == nil, "expected nil RawFile")
}
