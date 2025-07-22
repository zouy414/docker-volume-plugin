package utils

import (
	"fmt"
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
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("bind failed: %v, output: %s", err, string(output))
	}
	return nil
}

// MountNFS mounts an NFS share to a local path.
func MountNFS(address string, remotePath string, localPath string, mountOptions []string) error {
	if len(mountOptions) == 0 {
		mountOptions = []string{"defaults"}
	}

	cmd := exec.Command("mount", "-t", "nfs", "-o", strings.Join(mountOptions, ","), fmt.Sprintf("%s:%s", address, remotePath), localPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mount failed: %v, output: %s", err, string(output))
	}
	return nil
}

// UmountNFS unmounts an NFS share from a local path.
func Umount(localPath string) error {
	cmd := exec.Command("umount", localPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("umount failed: %v, output: %s", err, string(output))
	}
	return nil
}

// isMounted check if a local path is mount point.
func IsMounted(path string) (bool, error) {
	return mountinfo.Mounted(path)
}
