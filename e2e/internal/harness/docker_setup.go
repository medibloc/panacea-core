package harness

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/errdefs"
	"github.com/strangelove-ventures/interchaintest/v8/dockerutil"
)

const (
	dockerSetupTimeout          = 45 * time.Second
	dockerSubnetCandidates      = 16 * 256
	dockerNetworkCreateAttempts = 8
)

type dockerSetupClient interface {
	dockerResourceClient
	NetworkCreate(context.Context, string, dockertypes.NetworkCreate) (dockertypes.NetworkCreateResponse, error)
}

func setupDockerNetwork(parent context.Context, client dockerSetupClient, runID string) (string, error) {
	if parent == nil {
		return "", errors.New("Docker setup context is required")
	}
	if strings.TrimSpace(runID) == "" {
		return "", errors.New("Docker setup run ID is required")
	}

	ctx, cancel := context.WithTimeout(parent, dockerSetupTimeout)
	defer cancel()

	if err := cleanupDockerResourcesWithContext(ctx, client, runID); err != nil {
		return "", fmt.Errorf("clean stale Docker resources for %s: %w", runID, err)
	}

	start, err := randomDockerSubnetStart()
	if err != nil {
		return "", err
	}

	name := fmt.Sprintf("%s-%s", dockerutil.ICTDockerPrefix, dockerutil.RandLowerCaseLetterString(8))
	var (
		rejectedSubnets []*net.IPNet
		lastOverlapErr  error
	)
	for attempt := 1; attempt <= dockerNetworkCreateAttempts; attempt++ {
		usedSubnets, err := listUsedDockerSubnets(ctx, client)
		if err != nil {
			return "", err
		}
		usedSubnets = append(usedSubnets, rejectedSubnets...)
		subnet, err := selectAvailableDockerSubnet(usedSubnets, start)
		if err != nil {
			return "", err
		}

		created, err := boundedDockerResult(ctx, func(operationCtx context.Context) (dockertypes.NetworkCreateResponse, error) {
			return client.NetworkCreate(operationCtx, name, dockertypes.NetworkCreate{
				CheckDuplicate: true,
				Driver:         "bridge",
				IPAM: &network.IPAM{Config: []network.IPAMConfig{{
					Subnet: subnet,
				}}},
				Labels: map[string]string{dockerutil.CleanupLabel: runID},
			})
		})
		if err == nil {
			if err := ctx.Err(); err != nil {
				return "", fmt.Errorf("create Docker network for %s: %w", runID, err)
			}
			if strings.TrimSpace(created.ID) == "" {
				return "", fmt.Errorf("create Docker network for %s: Docker returned an empty network ID", runID)
			}
			return created.ID, nil
		}
		if !isDockerSubnetOverlapError(err) {
			return "", fmt.Errorf("create Docker network for %s: %w", runID, err)
		}

		lastOverlapErr = err
		_, rejected, parseErr := net.ParseCIDR(subnet)
		if parseErr != nil {
			return "", fmt.Errorf("record rejected Docker subnet %q: %w", subnet, parseErr)
		}
		rejectedSubnets = append(rejectedSubnets, rejected)
	}
	return "", fmt.Errorf(
		"create Docker network for %s after %d subnet-overlap attempts: %w",
		runID,
		dockerNetworkCreateAttempts,
		lastOverlapErr,
	)
}

func isDockerSubnetOverlapError(err error) bool {
	return errdefs.IsForbidden(err) && strings.Contains(strings.ToLower(err.Error()), "pool overlaps")
}

func listUsedDockerSubnets(ctx context.Context, client dockerSetupClient) ([]*net.IPNet, error) {
	networks, err := boundedDockerResult(ctx, func(operationCtx context.Context) ([]dockertypes.NetworkResource, error) {
		return client.NetworkList(operationCtx, dockertypes.NetworkListOptions{})
	})
	if err != nil {
		return nil, fmt.Errorf("list Docker networks for subnet selection: %w", err)
	}
	var used []*net.IPNet
	for _, item := range networks {
		for _, config := range item.IPAM.Config {
			if strings.TrimSpace(config.Subnet) == "" {
				continue
			}
			_, parsed, err := net.ParseCIDR(config.Subnet)
			if err != nil {
				return nil, fmt.Errorf("parse Docker network %s subnet %q: %w", item.ID, config.Subnet, err)
			}
			used = append(used, parsed)
		}
	}
	return used, nil
}

func randomDockerSubnetStart() (int, error) {
	var random [2]byte
	if _, err := cryptorand.Read(random[:]); err != nil {
		return 0, fmt.Errorf("choose Docker subnet start: %w", err)
	}
	return int(binary.BigEndian.Uint16(random[:])) % dockerSubnetCandidates, nil
}

func selectAvailableDockerSubnet(used []*net.IPNet, start int) (string, error) {
	start %= dockerSubnetCandidates
	if start < 0 {
		start += dockerSubnetCandidates
	}
	for offset := 0; offset < dockerSubnetCandidates; offset++ {
		index := (start + offset) % dockerSubnetCandidates
		secondOctet := 16 + index/256
		thirdOctet := index % 256
		_, candidate, err := net.ParseCIDR(fmt.Sprintf("172.%d.%d.0/24", secondOctet, thirdOctet))
		if err != nil {
			return "", fmt.Errorf("construct Docker subnet candidate: %w", err)
		}
		if !dockerSubnetOverlaps(candidate, used) {
			return candidate.String(), nil
		}
	}
	return "", errors.New("no available Docker /24 subnet in 172.16.0.0/12")
}

func dockerSubnetOverlaps(candidate *net.IPNet, used []*net.IPNet) bool {
	for _, occupied := range used {
		if occupied == nil {
			continue
		}
		if candidate.Contains(occupied.IP) || occupied.Contains(candidate.IP) {
			return true
		}
	}
	return false
}
