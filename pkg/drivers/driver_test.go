package drivers

import (
	"context"
	"docker-volume-plugin/pkg/log"
	"fmt"
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	ctx := context.Background()
	logger := log.New("test-driver")

	_, err := New(ctx, logger, "invalid-driver", "/tmp/mock-mountpoint", "")
	if err == nil {
		t.Fatal("expected error for invalid driver, got nil")
	}
}

func TestDrivers(t *testing.T) {
	testMountpointFormatString := os.TempDir() + "/test-mountpoint-%s"

	cases := []struct {
		driver        string
		driverOptions string
	}{
		{
			driver:        "nfs",
			driverOptions: `{"address": "nfs-server.example.com", "remotePath": "/mock", "purgeAfterDelete": true, "allowMultipleMount": true, "mock": true}`,
		},
		{
			driver: "mock",
		},
	}
	defer func() {
		for _, c := range cases {
			_ = os.RemoveAll(fmt.Sprintf(testMountpointFormatString, c.driver))
		}
	}()

	for _, c := range cases {
		t.Run(c.driver, func(t *testing.T) {
			driver, err := New(context.Background(), log.New(c.driver), c.driver, fmt.Sprintf(testMountpointFormatString, c.driver), c.driverOptions)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if driver == nil {
				t.Fatal("expected driver instance, got nil")
			}

			// Test Create
			err = driver.Create("test", map[string]string{"purgeAfterDelete": "true"})
			if err != nil {
				t.Fatalf("got error when create volume test: %v", err)
			}

			// Test List
			volumeMetadataMap, err := driver.List()
			if err != nil {
				t.Fatalf("got error when list volume: %v", err)
			}
			if len(volumeMetadataMap) != 1 {
				t.Errorf("expected 1 volumes, got %d volume", len(volumeMetadataMap))
			}

			// Test Get
			_, err = driver.Get("test")
			if err != nil {
				t.Fatalf("got error when get volume: %v", err)
			}

			// Test Path
			mountpoint, err := driver.Path("test")
			if err != nil {
				t.Fatalf("got error when get path: %v", err)
			}
			if mountpoint == "" {
				t.Errorf("expected non-empty mountpoint for volume test, got empty string")
			}
			_, err = driver.Path("non-exist")
			if err == nil {
				t.Fatalf("expect got error when path not existed volume")
			}

			// Test Mount
			for _, id := range []string{"1", "2"} {
				_, err = driver.Mount("test", id)
				if err != nil {
					t.Fatalf("got error when mount volume with %s: %v", id, err)
				}
				_, err = driver.Mount("test", id)
				if err == nil {
					t.Fatalf("expect got error when mount mounted volume  with %s", id)
				}
			}
			_, err = driver.Mount("non-exist", "1")
			if err == nil {
				t.Fatalf("expect got error when mount not existed volume")
			}

			// Test Remove mounted volume
			err = driver.Remove("test")
			if err == nil {
				t.Fatalf("expect got error when remove mounted volume")
			}

			// Test Unmount
			for _, id := range []string{"1", "2"} {
				err = driver.Unmount("test", id)
				if err != nil {
					t.Fatalf("got error when umount volume by %s: %v", id, err)
				}
				err = driver.Unmount("test", id)
				if err == nil {
					t.Fatalf("expect got error when unmount unmounted volume with %s", id)
				}
			}
			err = driver.Unmount("non-exist", "1")
			if err == nil {
				t.Fatalf("expect got error when unmount not existed volume")
			}

			// Test Remove unmounted volume
			err = driver.Remove("test")
			if err != nil {
				t.Fatalf("got error when remove volume: %v", err)
			}

			// Test Remove non-existent volume
			err = driver.Remove("non-exist")
			if err == nil {
				t.Fatalf("expect got error when remove not existed volume")
			}

			// Test List after remove
			volumeMetadataMap, err = driver.List()
			if err != nil {
				t.Fatalf("got error when list volume: %v", err)
			}
			if len(volumeMetadataMap) != 0 {
				t.Errorf("expected 0 volumes, got %d volume ", len(volumeMetadataMap))
			}

			err = driver.Destroy()
			if err != nil {
				t.Errorf("expected no error on destroy for driver, got %v", err)
			}
		})
	}
}
