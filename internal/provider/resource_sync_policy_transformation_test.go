package provider

import (
	"testing"

	cybrapi "github.com/aharriscybr/terraform-provider-cybr-sh/internal/cyberark"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func Test_transformationValue(t *testing.T) {
	tests := []struct {
		name           string
		transformation types.String
		want           *cybrapi.TransformationValue
	}{{
		"Test transformationValue",
		types.StringValue("test_transformation"),
		&cybrapi.TransformationValue{Predefined: "test_transformation"},
	}, {
		"Test transformationValue with empty string",
		types.StringValue(""),
		&cybrapi.TransformationValue{Predefined: ""},
	}, {
		"Test transformationValue with nil",
		types.StringNull(),
		nil,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := transformationValue(tt.transformation)
			assert.Equal(t, tt.want, got, "transformationValue() = %v, want %v", got, tt.want)
		})
	}
}
