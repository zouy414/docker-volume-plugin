package utils

import "testing"

func TestIsMounted(t *testing.T) {
	mounted, err := IsMounted("/")
	if err != nil {
		t.Fatalf("got error when checking if / is mounted: %v", err)
	}
	if !mounted {
		t.Fatalf("/ should be a mount point")
	}

	mounted, err = IsMounted("/bin")
	if err != nil {
		t.Fatalf("got error when checking if /bin is mounted: %v", err)
	}
	if mounted {
		t.Fatalf("/bin should not be a mount point")
	}

	_, err = IsMounted("/non-exist")
	if err == nil {
		t.Fatalf("expected error when checking non-existent mount point, got nil")
	}
}
