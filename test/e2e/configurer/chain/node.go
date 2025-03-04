package chain

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/ingenuity-build/quicksilver/test/e2e/containers"
	"github.com/ingenuity-build/quicksilver/test/e2e/initialization"
)

type NodeConfig struct {
	initialization.Node

	OperatorAddress  string
	SnapshotInterval uint64
	chainID          string
	rpcClient        *rpchttp.HTTP
	t                *testing.T
	containerManager *containers.Manager

	// Add this to help with logging / tracking time since start.
	setupTime time.Time
}

// NewNodeConfig returns new initialized NodeConfig.
func NewNodeConfig(t *testing.T, initNode *initialization.Node, initConfig *initialization.NodeConfig, chainID string, containerManager *containers.Manager) *NodeConfig {
	return &NodeConfig{
		Node:             *initNode,
		SnapshotInterval: initConfig.SnapshotInterval,
		chainID:          chainID,
		containerManager: containerManager,
		t:                t,
		setupTime:        time.Now(),
	}
}

// Run runs a node container for the given nodeIndex.
// The node configuration must be already added to the chain config prior to calling this
// method.
func (n *NodeConfig) Run() error {
	n.t.Logf("starting node container: %s", n.Name)
	resource, err := n.containerManager.RunNodeResource(n.Name, n.ConfigDir)
	if err != nil {
		return err
	}

	hostPort := resource.GetHostPort("26657/tcp")
	rpcClient, err := rpchttp.New("tcp://"+hostPort, "/websocket")
	if err != nil {
		return err
	}

	n.rpcClient = rpcClient

	require.Eventually(
		n.t,
		func() bool {
			// This fails if unsuccessful.
			_, err := n.QueryCurrentHeight()
			if err != nil {
				return false
			}
			n.t.Logf("started node container: %s", n.Name)
			return true
		},
		2*time.Minute,
		time.Second,
		"Quicksilver node failed to produce blocks",
	)

	return n.extractOperatorAddressIfValidator()
}

// Stop stops the node from running and removes its container.
func (n *NodeConfig) Stop() error {
	n.t.Logf("stopping node container: %s", n.Name)
	if err := n.containerManager.RemoveNodeResource(n.Name); err != nil {
		return err
	}
	n.t.Logf("stopped node container: %s", n.Name)
	return nil
}

// WaitUntil waits until node reaches doneCondition. Return nil
// if reached, error otherwise.
func (n *NodeConfig) WaitUntil(doneCondition func(syncInfo coretypes.SyncInfo) bool) {
	var latestBlockHeight int64
	for i := 0; i < waitUntilrepeatMax; i++ {
		status, err := n.rpcClient.Status(context.Background())
		require.NoError(n.t, err)
		latestBlockHeight = status.SyncInfo.LatestBlockHeight
		// let the node produce a few blocks
		if !doneCondition(status.SyncInfo) {
			time.Sleep(waitUntilRepeatPauseTime)
			continue
		}
		return
	}
	n.t.Errorf("node %s timed out waiting for condition, latest block height was %d", n.Name, latestBlockHeight)
}

func (n *NodeConfig) extractOperatorAddressIfValidator() error {
	if !n.IsValidator {
		n.t.Logf("node (%s) is not a validator, skipping", n.Name)
		return nil
	}

	cmd := []string{"quicksilverd", "debug", "addr", n.PublicKey}
	n.t.Logf("extracting validator operator addresses for validator: %s", n.Name)
	_, errBuf, err := n.containerManager.ExecCmd(n.t, n.Name, cmd, "")
	if err != nil {
		return err
	}
	re := regexp.MustCompile("quickvaloper(.{39})")
	operAddr := fmt.Sprintf("%s\n", re.FindString(errBuf.String()))
	n.OperatorAddress = strings.TrimSuffix(operAddr, "\n")
	return nil
}

func (n *NodeConfig) GetHostPort(portID string) (string, error) {
	return n.containerManager.GetHostPort(n.Name, portID)
}

func (n *NodeConfig) WithSetupTime(t time.Time) *NodeConfig {
	n.setupTime = t
	return n
}

func (n *NodeConfig) LogActionF(msg string, args ...interface{}) {
	timeSinceStart := time.Since(n.setupTime).Round(time.Millisecond)
	s := fmt.Sprintf(msg, args...)
	n.t.Logf("[%s] %s. From container %s", timeSinceStart, s, n.Name)
}
