package cyberark

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPAMApi_filters(t *testing.T) {
	type args struct {
		search string
		filter []string
	}
	tests := []struct {
		name      string
		args      args
		wantQuery map[string]string
	}{
		{
			name:      "empty search and filter",
			args:      args{search: "", filter: []string{}},
			wantQuery: map[string]string{},
		},
		{
			name:      "search only",
			args:      args{search: "test%", filter: []string{}},
			wantQuery: map[string]string{"search": "test%"},
		},
		{
			name:      "filter only",
			args:      args{search: "", filter: []string{"test"}},
			wantQuery: map[string]string{"filter": "test"},
		},
		{
			name:      "search and filter",
			args:      args{search: "testSearch", filter: []string{"test filter"}},
			wantQuery: map[string]string{"filter": "test filter", "search": "testSearch"},
		},
		{
			name:      "multiple filters",
			args:      args{search: "", filter: []string{"test/", "test2"}},
			wantQuery: map[string]string{"filter": "test/ AND test2"},
		},
		{
			name:      "multiple filters and search",
			args:      args{search: "test", filter: []string{"test", "test2"}},
			wantQuery: map[string]string{"filter": "test AND test2", "search": "test"},
		},
		{
			name:      "search with special characters",
			args:      args{search: "test%25", filter: []string{}},
			wantQuery: map[string]string{"search": "test%25"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &pamAPI{}
			result := a.filters(tt.args.search, tt.args.filter)
			assert.Equalf(t, tt.wantQuery, result, "filters(%v, %v)", tt.args.search, tt.args.filter)
		})
	}
}
