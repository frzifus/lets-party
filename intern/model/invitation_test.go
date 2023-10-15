package model

import (
	"testing"

	"github.com/google/uuid"
)

func TestInvitation_RemoveGuest(t *testing.T) {
	tt := []struct{
		name string
		invite  Invitation
		toRemove uuid.UUID
	}{
		{
			name: "empty or notFound",
			invite: Invitation{
				GuestIDs: []uuid.UUID{},
			},
			toRemove: uuid.MustParse("b5627acd-9332-476c-8466-f49de1567865"),
		},
		{
			name: "remove first",
			invite: Invitation{
				GuestIDs: []uuid.UUID{
					uuid.MustParse("0eac703a-40f3-4318-ae96-f28e026a23c6"),
					uuid.MustParse("b5627acd-9332-476c-8466-f49de1567865"),
					uuid.MustParse("951812f2-9bbd-481b-a798-6653c355b9c0"),
				},
			},
			toRemove: uuid.MustParse("0eac703a-40f3-4318-ae96-f28e026a23c6"),
		},
		{
			name: "remove last",
			invite: Invitation{
				GuestIDs: []uuid.UUID{
					uuid.MustParse("0eac703a-40f3-4318-ae96-f28e026a23c6"),
					uuid.MustParse("b5627acd-9332-476c-8466-f49de1567865"),
					uuid.MustParse("951812f2-9bbd-481b-a798-6653c355b9c0"),
				},
			},
			toRemove: uuid.MustParse("951812f2-9bbd-481b-a798-6653c355b9c0"),
		},
		{
			name: "remove mid",
			invite: Invitation{
				GuestIDs: []uuid.UUID{
					uuid.MustParse("0eac703a-40f3-4318-ae96-f28e026a23c6"),
					uuid.MustParse("b5627acd-9332-476c-8466-f49de1567865"),
					uuid.MustParse("951812f2-9bbd-481b-a798-6653c355b9c0"),
				},
			},
			toRemove: uuid.MustParse("b5627acd-9332-476c-8466-f49de1567865"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tc.invite.RemoveGuest(tc.toRemove)
			for _, invID := range tc.invite.GuestIDs {
				if invID == tc.toRemove {
					t.Fatalf("guestID still exists: %s", invID.String())
				}
			}
		})
	}
}
