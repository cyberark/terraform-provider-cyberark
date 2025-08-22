package cyberark_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/cyberark/terraform-provider-cyberark/internal/cyberark"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestValidateInputField(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		val         types.String
		fieldName   string
		minLen      int
		maxLen      int
		pattern     string
		expectError bool
		errorSubstr string
	}{
		{
			name:        "Null value",
			val:         types.StringNull(),
			fieldName:   "test_field",
			expectError: false,
		},
		{
			name:        "Unknown value",
			val:         types.StringUnknown(),
			fieldName:   "test_field",
			expectError: false,
		},
		{
			name:        "Valid value",
			val:         types.StringValue("Valid"),
			fieldName:   "test_field",
			minLen:      1,
			maxLen:      10,
			pattern:     "^[a-zA-Z]+$",
			expectError: false,
		},
		{
			name:        "Too short",
			val:         types.StringValue("a"),
			fieldName:   "test_field",
			minLen:      5,
			maxLen:      10,
			pattern:     "^[a-zA-Z]+$",
			expectError: true,
			errorSubstr: `must be between 5 and 10 characters`,
		},
		{
			name:        "Too long",
			val:         types.StringValue("thisisaverylongstring"),
			fieldName:   "test_field",
			minLen:      1,
			maxLen:      5,
			pattern:     "^[a-zA-Z]+$",
			expectError: true,
			errorSubstr: `must be between 1 and 5 characters`,
		},
		{
			name:        "Pattern mismatch",
			val:         types.StringValue("valid_string"),
			fieldName:   "test_field",
			minLen:      1,
			maxLen:      20,
			pattern:     "^[0-9]+$",
			expectError: true,
			errorSubstr: `must match the pattern`,
		},
		{
			name:        "Regex compilation error",
			val:         types.StringValue("valid"),
			fieldName:   "test_field",
			minLen:      1,
			maxLen:      10,
			pattern:     "(((",
			expectError: true,
			errorSubstr: "regex pattern error",
		},		
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cyberark.ValidateInputField(ctx, tt.fieldName, tt.val, tt.minLen, tt.maxLen, tt.pattern)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorSubstr, fmt.Sprintf("Expected error to contain %q", tt.errorSubstr))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}