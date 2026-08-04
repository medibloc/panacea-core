package harness

import (
	"errors"
	"sync"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/require"
)

func TestNewUpgradeSwitchPlan(t *testing.T) {
	oldImage := ibc.DockerImage{Repository: "panacea-e2e-v2.2.1", Version: "local", UIDGID: "0:0"}
	plan, err := newUpgradeSwitchPlan(
		"panacea-validator-2",
		oldImage,
		ImageRef{Repository: "panacea-e2e-current", Version: "local"},
	)
	require.NoError(t, err)
	require.Equal(t, "panacea-validator-2", plan.Node)
	require.Equal(t, oldImage, plan.From)
	require.Equal(t, "panacea-e2e-current", plan.To.Repository)
	require.Equal(t, "local", plan.To.Version)
	require.Equal(t, oldImage.UIDGID, plan.To.UIDGID)

	_, err = newUpgradeSwitchPlan("", oldImage, ImageRef{Repository: "current", Version: "local"})
	require.ErrorContains(t, err, "node name")
	_, err = newUpgradeSwitchPlan("node", oldImage, ImageRef{})
	require.ErrorContains(t, err, "target image")
	_, err = newUpgradeSwitchPlan("node", oldImage, ImageRef{Repository: oldImage.Repository, Version: oldImage.Version})
	require.ErrorContains(t, err, "already uses")
}

func TestRunUpgradeSwitchOperationsRemovesOldContainerBeforeCreatingNewOne(t *testing.T) {
	var calls []string
	operation := func(name string) func() error {
		return func() error {
			calls = append(calls, name)
			return nil
		}
	}
	err := runUpgradeSwitchOperations(true, upgradeSwitchOperations{
		capture: operation("capture"),
		stop:    operation("stop"),
		remove:  operation("remove"),
		create:  operation("create"),
		start:   operation("start"),
	})
	require.NoError(t, err)
	require.Equal(t, []string{"capture", "stop", "remove", "create", "start"}, calls)

	calls = nil
	err = runUpgradeSwitchOperations(false, upgradeSwitchOperations{
		capture: operation("capture"),
		stop:    operation("stop"),
		remove:  operation("remove"),
		create:  operation("create"),
		start:   operation("start"),
	})
	require.NoError(t, err)
	require.Equal(t, []string{"capture", "remove", "create", "start"}, calls)

	calls = nil
	err = runUpgradeSwitchOperations(false, upgradeSwitchOperations{
		capture: operation("capture"),
		stop:    operation("stop"),
		remove:  operation("remove"),
		create: func() error {
			calls = append(calls, "create")
			return errors.New("create failed")
		},
		start: operation("start"),
	})
	require.ErrorContains(t, err, "create failed")
	require.Equal(t, []string{"capture", "remove", "create"}, calls)
}

func TestRunUpgradeBatchSwitchOperationsCreatesEveryContainerBeforeConcurrentStart(t *testing.T) {
	const nodeCount = 4
	var mu sync.Mutex
	created := 0
	started := 0
	allStarted := make(chan struct{})
	switches := make([]upgradeBatchSwitch, nodeCount)
	for index := range switches {
		switches[index] = upgradeBatchSwitch{operations: upgradeSwitchOperations{
			capture: func() error { return nil },
			remove:  func() error { return nil },
			create: func() error {
				mu.Lock()
				created++
				mu.Unlock()
				return nil
			},
			start: func() error {
				mu.Lock()
				defer mu.Unlock()
				if created != nodeCount {
					return errors.New("start ran before every replacement was created")
				}
				started++
				if started == nodeCount {
					close(allStarted)
				}
				return nil
			},
		}}
	}

	results := runUpgradeBatchSwitchOperations(switches)
	require.Len(t, results, nodeCount)
	for _, err := range results {
		require.NoError(t, err)
	}
	<-allStarted
	require.Equal(t, nodeCount, created)
	require.Equal(t, nodeCount, started)
}
