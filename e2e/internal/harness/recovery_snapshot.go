package harness

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// LocalSnapshot is one snapshot advertised by `panacead snapshots list`.
type LocalSnapshot struct {
	Height uint64 `json:"height"`
	Format uint32 `json:"format"`
	Chunks uint32 `json:"chunks"`
}

// ParseLocalSnapshots parses the stable, line-oriented output of the Cosmos
// SDK snapshot command. Unknown non-empty output is rejected so a CLI change
// cannot silently turn restore coverage into a no-op.
func ParseLocalSnapshots(output []byte) ([]LocalSnapshot, error) {
	var snapshots []LocalSnapshot
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 6 || fields[0] != "height:" || fields[2] != "format:" || fields[4] != "chunks:" {
			return nil, fmt.Errorf("unrecognized snapshot list line %q", line)
		}
		height, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil || height == 0 {
			return nil, fmt.Errorf("invalid snapshot height %q", fields[1])
		}
		format, err := strconv.ParseUint(fields[3], 10, 32)
		if err != nil || format == 0 {
			return nil, fmt.Errorf("invalid snapshot format %q", fields[3])
		}
		chunks, err := strconv.ParseUint(fields[5], 10, 32)
		if err != nil || chunks == 0 {
			return nil, fmt.Errorf("invalid snapshot chunk count %q", fields[5])
		}
		snapshots = append(snapshots, LocalSnapshot{
			Height: height,
			Format: uint32(format),
			Chunks: uint32(chunks),
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read snapshot list: %w", err)
	}
	return snapshots, nil
}

// FindLocalSnapshot returns the unique snapshot at height.
func FindLocalSnapshot(snapshots []LocalSnapshot, height uint64) (LocalSnapshot, error) {
	var found *LocalSnapshot
	for i := range snapshots {
		if snapshots[i].Height != height {
			continue
		}
		if found != nil {
			return LocalSnapshot{}, fmt.Errorf("multiple local snapshots exist at height %d", height)
		}
		candidate := snapshots[i]
		found = &candidate
	}
	if found == nil {
		return LocalSnapshot{}, fmt.Errorf("local snapshot at height %d was not listed", height)
	}
	return *found, nil
}
