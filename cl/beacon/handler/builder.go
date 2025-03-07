package handler

import (
	"net/http"

	libcommon "github.com/ledgerwatch/erigon-lib/common"
	"github.com/ledgerwatch/erigon/cl/beacon/beaconhttp"
	"github.com/ledgerwatch/erigon/cl/clparams"
	"github.com/ledgerwatch/erigon/cl/persistence/beacon_indicies"
	"github.com/ledgerwatch/erigon/cl/phase1/core/state"
)

func (a *ApiHandler) GetEth1V1BuilderStatesExpectedWit(w http.ResponseWriter, r *http.Request) (*beaconResponse, error) {
	ctx := r.Context()

	tx, err := a.indiciesDB.BeginRo(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	blockId, err := stateIdFromRequest(r)
	if err != nil {
		return nil, beaconhttp.NewEndpointError(http.StatusBadRequest, err.Error())
	}
	root, httpStatus, err := a.blockRootFromStateId(ctx, tx, blockId)
	if err != nil {
		return nil, beaconhttp.NewEndpointError(httpStatus, err.Error())
	}
	slot, err := beacon_indicies.ReadBlockSlotByBlockRoot(tx, root)
	if err != nil {
		return nil, err
	}
	if slot == nil {
		return nil, beaconhttp.NewEndpointError(http.StatusNotFound, "state not found")
	}
	if a.beaconChainCfg.GetCurrentStateVersion(*slot/a.beaconChainCfg.SlotsPerEpoch) < clparams.CapellaVersion {
		return nil, beaconhttp.NewEndpointError(http.StatusBadRequest, "the specified state is not a capella state")
	}
	headRoot, _, err := a.forkchoiceStore.GetHead()
	if err != nil {
		return nil, err
	}
	if root == headRoot {
		s, cn := a.syncedData.HeadState()
		defer cn()
		return newBeaconResponse(state.ExpectedWithdrawals(s)).withFinalized(false), nil
	}
	lookAhead := 1024
	for currSlot := *slot + 1; currSlot < *slot+uint64(lookAhead); currSlot++ {
		if currSlot > a.syncedData.HeadSlot() {
			return nil, beaconhttp.NewEndpointError(http.StatusNotFound, "state not found")
		}
		blockRoot, err := beacon_indicies.ReadCanonicalBlockRoot(tx, currSlot)
		if err != nil {
			return nil, err
		}
		if blockRoot == (libcommon.Hash{}) {
			continue
		}
		blk, err := a.blockReader.ReadBlockByRoot(ctx, tx, blockRoot)
		if err != nil {
			return nil, err
		}
		return newBeaconResponse(blk.Block.Body.ExecutionPayload.Withdrawals).withFinalized(false), nil
	}

	return nil, beaconhttp.NewEndpointError(http.StatusNotFound, "state not found")
}
