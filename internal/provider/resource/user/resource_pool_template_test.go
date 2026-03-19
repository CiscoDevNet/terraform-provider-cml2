package user

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestResourcePoolTemplateChanged(t *testing.T) {
	for _, tc := range []struct {
		name  string
		plan  types.String
		state types.String
		want  bool
	}{
		{
			name:  "both null",
			plan:  types.StringNull(),
			state: types.StringNull(),
			want:  false,
		},
		{
			name:  "same value",
			plan:  types.StringValue("00000000-0000-4000-8000-000000000001"),
			state: types.StringValue("00000000-0000-4000-8000-000000000001"),
			want:  false,
		},
		{
			name:  "changed value",
			plan:  types.StringValue("00000000-0000-4000-8000-000000000001"),
			state: types.StringValue("00000000-0000-4000-8000-000000000002"),
			want:  true,
		},
		{
			name:  "state set, plan null",
			plan:  types.StringNull(),
			state: types.StringValue("00000000-0000-4000-8000-000000000001"),
			want:  true,
		},
		{
			name:  "state null, plan set",
			plan:  types.StringValue("00000000-0000-4000-8000-000000000001"),
			state: types.StringNull(),
			want:  true,
		},
		{
			name:  "plan unknown",
			plan:  types.StringUnknown(),
			state: types.StringValue("00000000-0000-4000-8000-000000000001"),
			want:  false,
		},
		{
			name:  "state unknown",
			plan:  types.StringValue("00000000-0000-4000-8000-000000000001"),
			state: types.StringUnknown(),
			want:  false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := resourcePoolTemplateChanged(tc.plan, tc.state); got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}
