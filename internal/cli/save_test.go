package cli

import "testing"

func TestValidateSaveTarget(t *testing.T) {
	tests := []struct {
		name      string
		title     string
		parentUID string
		pageUID   string
		wantErr   bool
	}{
		{name: "page mode ok", title: "Page", wantErr: false},
		{name: "parent mode ok", parentUID: "p1", wantErr: false},
		{name: "missing both", wantErr: true},
		{name: "both set", title: "Page", parentUID: "p1", wantErr: true},
		{name: "uid with parent", parentUID: "p1", pageUID: "x", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSaveTarget(tt.title, tt.parentUID, tt.pageUID)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
