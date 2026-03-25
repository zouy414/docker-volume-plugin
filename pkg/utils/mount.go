package utils

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/moby/sys/mountinfo"
)

// Bind mounts a directory to a directory.
func Bind(sourcePath string, directoryPath string, mountOptions []string) error {
	if len(mountOptions) == 0 {
		mountOptions = []string{"defaults"}
	}

	cmd := exec.Command("mount", "--bind", "-o", strings.Join(mountOptions, ","), sourcePath, directoryPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("bind failed: %v, output: %s", err, string(output))
	}

	return nil
}

// MountNFS mounts an NFS share to a local path.
func MountNFS(address string, remotePath string, localPath string, mountOptions []string) error {
	// Create the mount point if it doesn't exist
	if err := os.MkdirAll(localPath, 0755); err != nil {
		return fmt.Errorf("failed to create mount point: %v", err)
	}

	if len(mountOptions) == 0 {
		mountOptions = []string{"defaults"}
	}

	cmd := exec.Command("mount", "-t", "nfs", "-o", strings.Join(mountOptions, ","), fmt.Sprintf("%s:%s", address, remotePath), localPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("mount failed: %v, output: %s", err, string(output))
	}

	return nil
}

// MountCIFS mounts a CIFS share to a local path.
func MountCIFS(address string, remotePath string, localPath string, username string, password string, mountOptions []string) error {
	// Create the mount point if it doesn't exist
	if err := os.MkdirAll(localPath, 0755); err != nil {
		return fmt.Errorf("failed to create mount point: %v", err)
	}

	mountOptionsString := "username=" + username
	if len(password) != 0 {
		mountOptionsString = mountOptionsString + ",password=" + password
	}
	if len(mountOptions) != 0 {
		mountOptionsString = mountOptionsString + "," + strings.Join(mountOptions, ",")
	}

	cmd := exec.Command("mount", "-t", "cifs", "-o", mountOptionsString, fmt.Sprintf("//%s%s", address, remotePath), localPath)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("mount failed: %v, output: %s, %s", err, string(output), fmt.Sprintf("//%s%s", address, remotePath))
	}

	return nil
}

// MountMock simulates mounting by creating the mount point directory without performing an actual mount.
func MountMock(localPath string) error {
	// Create the mount point if it doesn't exist
	if err := os.MkdirAll(localPath, 0755); err != nil {
		return fmt.Errorf("failed to create mount point: %v", err)
	}

	return nil
}

// Umount a volume from a local path.
func Umount(localPath string) error {
	cmd := exec.Command("umount", localPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("umount failed: %v, output: %s", err, string(output))
	}
	return nil
}

// IsMounted check if a local path is mount point.
func IsMounted(path string) (bool, error) {
	return mountinfo.Mounted(path)
}
