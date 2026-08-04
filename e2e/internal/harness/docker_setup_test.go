package harness

import (
	"context"
	"errors"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	volumetypes "github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/errdefs"
	"github.com/strangelove-ventures/interchaintest/v8/dockerutil"
	"github.com/stretchr/testify/require"
)

type fakeDockerSetupClient struct {
	calls             []string
	missingDeadline   bool
	excessiveDeadline bool
	blockCall         string
	containerListCall int
	volumeListCall    int
	networkListCall   int
	networkCreateCall int
	createName        string
	createOptions     dockertypes.NetworkCreate
	createdSubnets    []string
	createErrors      []error
	createError       error
}

func (f *fakeDockerSetupClient) record(ctx context.Context, call string) error {
	f.calls = append(f.calls, call)
	deadline, ok := ctx.Deadline()
	if !ok {
		f.missingDeadline = true
	} else if time.Until(deadline) > 11*time.Second {
		f.excessiveDeadline = true
	}
	if f.blockCall == call {
		<-ctx.Done()
		return ctx.Err()
	}
	return nil
}

func (f *fakeDockerSetupClient) ContainerList(ctx context.Context, _ dockertypes.ContainerListOptions) ([]dockertypes.Container, error) {
	f.containerListCall++
	call := "container-list-" + strconv.Itoa(f.containerListCall)
	if err := f.record(ctx, call); err != nil {
		return nil, err
	}
	if f.containerListCall == 1 {
		return []dockertypes.Container{{ID: "stale-container"}}, nil
	}
	return nil, nil
}

func (f *fakeDockerSetupClient) ContainerStop(ctx context.Context, _ string, _ container.StopOptions) error {
	return f.record(ctx, "container-stop")
}

func (f *fakeDockerSetupClient) ContainerRemove(ctx context.Context, _ string, _ dockertypes.ContainerRemoveOptions) error {
	return f.record(ctx, "container-remove")
}

func (f *fakeDockerSetupClient) VolumeList(ctx context.Context, _ volumetypes.ListOptions) (volumetypes.ListResponse, error) {
	f.volumeListCall++
	call := "volume-list-" + strconv.Itoa(f.volumeListCall)
	if err := f.record(ctx, call); err != nil {
		return volumetypes.ListResponse{}, err
	}
	if f.volumeListCall == 1 {
		return volumetypes.ListResponse{Volumes: []*volumetypes.Volume{{Name: "stale-volume"}}}, nil
	}
	return volumetypes.ListResponse{}, nil
}

func (f *fakeDockerSetupClient) VolumeRemove(ctx context.Context, _ string, _ bool) error {
	return f.record(ctx, "volume-remove")
}

func (f *fakeDockerSetupClient) NetworkList(ctx context.Context, _ dockertypes.NetworkListOptions) ([]dockertypes.NetworkResource, error) {
	f.networkListCall++
	call := "network-list-" + strconv.Itoa(f.networkListCall)
	if err := f.record(ctx, call); err != nil {
		return nil, err
	}
	switch f.networkListCall {
	case 1:
		return []dockertypes.NetworkResource{{ID: "stale-network"}}, nil
	case 3:
		return []dockertypes.NetworkResource{{
			ID:   "used-network",
			IPAM: network.IPAM{Config: []network.IPAMConfig{{Subnet: "172.16.0.0/24"}}},
		}}, nil
	default:
		return nil, nil
	}
}

func (f *fakeDockerSetupClient) NetworkRemove(ctx context.Context, _ string) error {
	return f.record(ctx, "network-remove")
}

func (f *fakeDockerSetupClient) NetworkCreate(ctx context.Context, name string, options dockertypes.NetworkCreate) (dockertypes.NetworkCreateResponse, error) {
	f.networkCreateCall++
	call := "network-create-" + strconv.Itoa(f.networkCreateCall)
	if err := f.record(ctx, call); err != nil {
		return dockertypes.NetworkCreateResponse{}, err
	}
	f.createName = name
	f.createOptions = options
	f.createdSubnets = append(f.createdSubnets, options.IPAM.Config[0].Subnet)
	if f.networkCreateCall <= len(f.createErrors) && f.createErrors[f.networkCreateCall-1] != nil {
		return dockertypes.NetworkCreateResponse{}, f.createErrors[f.networkCreateCall-1]
	}
	if f.createError != nil {
		return dockertypes.NetworkCreateResponse{}, f.createError
	}
	return dockertypes.NetworkCreateResponse{ID: "created-network"}, nil
}

func TestDockerSetupUsesLongerOverallBudgetThanEachOperation(t *testing.T) {
	require.Equal(t, 45*time.Second, dockerSetupTimeout)
	require.Equal(t, 10*time.Second, dockerOperationTimeout)
	require.Greater(t, dockerSetupTimeout, dockerOperationTimeout)
}

func TestSetupDockerNetworkIsBoundedAndCleansBeforeCreate(t *testing.T) {
	client := &fakeDockerSetupClient{}
	networkID, err := setupDockerNetwork(context.Background(), client, "run-bounded")
	require.NoError(t, err)
	require.Equal(t, "created-network", networkID)
	require.False(t, client.missingDeadline)
	require.False(t, client.excessiveDeadline)
	require.Equal(t, []string{
		"container-list-1", "container-stop", "container-remove",
		"volume-list-1", "volume-remove",
		"network-list-1", "network-remove",
		"container-list-2", "volume-list-2", "network-list-2",
		"network-list-3", "network-create-1",
	}, client.calls)
	require.True(t, strings.HasPrefix(client.createName, dockerutil.ICTDockerPrefix+"-"))
	require.True(t, client.createOptions.CheckDuplicate)
	require.Equal(t, "bridge", client.createOptions.Driver)
	require.Equal(t, "run-bounded", client.createOptions.Labels[dockerutil.CleanupLabel])
	require.NotNil(t, client.createOptions.IPAM)
	require.Len(t, client.createOptions.IPAM.Config, 1)
	_, selected, err := net.ParseCIDR(client.createOptions.IPAM.Config[0].Subnet)
	require.NoError(t, err)
	require.Equal(t, 24, maskSize(t, selected))
	require.False(t, selected.Contains(net.ParseIP("172.16.0.1")))
}

func TestSetupDockerNetworkHonorsCallerDeadline(t *testing.T) {
	for _, blockCall := range []string{"container-list-1", "network-list-3", "network-create-1"} {
		t.Run(blockCall, func(t *testing.T) {
			client := &fakeDockerSetupClient{blockCall: blockCall}
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
			defer cancel()
			started := time.Now()
			_, err := setupDockerNetwork(ctx, client, "run-timeout")
			require.ErrorIs(t, err, context.DeadlineExceeded)
			require.Less(t, time.Since(started), time.Second)
			require.False(t, client.missingDeadline)
		})
	}
}

func TestSetupDockerNetworkRetriesSubnetOverlapWithFreshListAndSubnet(t *testing.T) {
	overlapErr := errdefs.Forbidden(errors.New("Pool overlaps with other one on this address space"))
	client := &fakeDockerSetupClient{createErrors: []error{overlapErr}}

	networkID, err := setupDockerNetwork(context.Background(), client, "run-overlap")
	require.NoError(t, err)
	require.Equal(t, "created-network", networkID)
	require.Equal(t, 2, client.networkCreateCall)
	require.Equal(t, 4, client.networkListCall)
	require.Len(t, client.createdSubnets, 2)
	require.NotEqual(t, client.createdSubnets[0], client.createdSubnets[1])
	require.Equal(t, []string{
		"network-list-3", "network-create-1",
		"network-list-4", "network-create-2",
	}, client.calls[len(client.calls)-4:])
}

func TestSetupDockerNetworkBoundsSubnetOverlapRetries(t *testing.T) {
	const expectedAttempts = 8
	overlapErr := errdefs.Forbidden(errors.New("Address pool overlaps with existing pool on this address space"))
	client := &fakeDockerSetupClient{createError: overlapErr}

	_, err := setupDockerNetwork(context.Background(), client, "run-overlap-limit")
	require.ErrorIs(t, err, overlapErr)
	require.ErrorContains(t, err, "subnet-overlap attempts")
	require.Equal(t, expectedAttempts, client.networkCreateCall)
	require.Equal(t, expectedAttempts+2, client.networkListCall)
	require.Len(t, client.createdSubnets, expectedAttempts)
	require.Len(t, uniqueStrings(client.createdSubnets), expectedAttempts)
}

func TestSetupDockerNetworkDoesNotRetryOtherForbiddenErrors(t *testing.T) {
	permissionErr := errdefs.Forbidden(errors.New("permission denied"))
	client := &fakeDockerSetupClient{createError: permissionErr}

	_, err := setupDockerNetwork(context.Background(), client, "run-forbidden")
	require.ErrorIs(t, err, permissionErr)
	require.NotContains(t, err.Error(), "subnet-overlap attempts")
	require.Equal(t, 1, client.networkCreateCall)
	require.Equal(t, 3, client.networkListCall)
}

func TestSelectAvailableDockerSubnetTerminatesWhenExhausted(t *testing.T) {
	_, used, err := net.ParseCIDR("172.16.0.0/12")
	require.NoError(t, err)
	_, err = selectAvailableDockerSubnet([]*net.IPNet{used}, 0)
	require.ErrorContains(t, err, "no available Docker /24 subnet")
}

func maskSize(t *testing.T, network *net.IPNet) int {
	t.Helper()
	ones, bits := network.Mask.Size()
	require.Equal(t, 32, bits)
	return ones
}

func uniqueStrings(values []string) map[string]struct{} {
	unique := make(map[string]struct{}, len(values))
	for _, value := range values {
		unique[value] = struct{}{}
	}
	return unique
}
