package handler

import (
	"encoding/json"
	"net/http"

	"github.com/ledgerwatch/erigon/cl/beacon/beaconhttp"
)

func (a *ApiHandler) GetEthV2DebugBeaconHeads(w http.ResponseWriter, r *http.Request) (*beaconResponse, error) {
	if a.syncedData.Syncing() {
		return nil, beaconhttp.NewEndpointError(http.StatusServiceUnavailable, "beacon node is syncing")
	}
	hash, slotNumber, err := a.forkchoiceStore.GetHead()
	if err != nil {
		return nil, err
	}
	return newBeaconResponse(
		[]interface{}{
			map[string]interface{}{
				"slot":                 slotNumber,
				"root":                 hash,
				"execution_optimistic": false,
			},
		}), nil
}

func (a *ApiHandler) GetEthV1DebugBeaconForkChoice(w http.ResponseWriter, r *http.Request) {
	justifiedCheckpoint := a.forkchoiceStore.JustifiedCheckpoint()
	finalizedCheckpoint := a.forkchoiceStore.FinalizedCheckpoint()
	forkNodes := a.forkchoiceStore.ForkNodes()
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"justified_checkpoint": justifiedCheckpoint,
		"finalized_checkpoint": finalizedCheckpoint,
		"fork_choice_nodes":    forkNodes,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
