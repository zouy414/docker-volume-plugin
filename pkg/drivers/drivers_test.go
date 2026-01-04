package drivers

import (
	"context"
	"docker-volume-plugin/pkg/log"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	ctx := context.Background()
	logger := log.New("test")

	_, err := New(ctx, logger, "invalid-driver", "/tmp/mock-mountpoint", "")
	assert.Error(t, err)

	driver, err := New(ctx, logger, "mock", "/tmp/mock-mountpoint", "")
	assert.NoError(t, err)
	assert.NotNil(t, driver)
}

func TestDrivers(t *testing.T) {
	testMountpointFormatString := os.TempDir() + "/test-mountpoint-%s"

	cases := []struct {
		driver        string
		driverOptions string
	}{
		{
			driver: "mock",
		},
		{
			driver:        "nfs",
			driverOptions: `{"address": "nfs-server.example.com", "remotePath": "/mock", "mock": true}`,
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
			assert.NoError(t, err)
			assert.NotNil(t, driver)

			// Test Create
			err = driver.Create("test", map[string]string{"purgeAfterDelete": "true"})
			assert.NoError(t, err)

			// Test List
			volumeMetadataMap, err := driver.List()
			assert.NoError(t, err)
			assert.NotEmpty(t, volumeMetadataMap)

			// Test Get
			_, err = driver.Get("test")
			assert.NoError(t, err)

			// Test Path exist Volume
			mountpoint, err := driver.Path("test")
			assert.NoError(t, err)
			assert.NotEmpty(t, mountpoint)

			// Test Path non-exist Volume
			_, err = driver.Path("non-exist")
			assert.Error(t, err)

			// Test Mount
			_, err = driver.Mount("test", "4103b9f9-189c-4a12-b1fb-5511ddc18297")
			assert.NoError(t, err)

			// Test Mount already mounted volume
			_, err = driver.Mount("test", "4103b9f9-189c-4a12-b1fb-5511ddc18297")
			assert.Error(t, err)

			// Test Mount non-exist volume
			_, err = driver.Mount("non-exist", "4103b9f9-189c-4a12-b1fb-5511ddc18297")
			assert.Error(t, err)

			// Test Unmount
			err = driver.Unmount("test", "4103b9f9-189c-4a12-b1fb-5511ddc18297")
			assert.NoError(t, err)

			// Test Unmount non-mounted volume
			err = driver.Unmount("test", "4103b9f9-189c-4a12-b1fb-5511ddc18297")
			assert.NoError(t, err)

			// Test Unmount non-exist volume
			err = driver.Unmount("non-exist", "4103b9f9-189c-4a12-b1fb-5511ddc18297")
			assert.Error(t, err)

			// Test Remove volume
			err = driver.Remove("test")
			assert.NoError(t, err)

			// Test Remove non-existent volume
			err = driver.Remove("non-exist")
			assert.Error(t, err)

			// Test List after remove
			volumeMetadataMap, err = driver.List()
			assert.NoError(t, err)
			assert.Empty(t, volumeMetadataMap)

			// Test Destroy
			err = driver.Destroy()
			assert.NoError(t, err)
		})
	}
}
