package iavlviewer

import (
	"errors"
	"fmt"

	dbm "github.com/cometbft/cometbft-db"
	"github.com/cosmos/cosmos-sdk/store/types"
)

// From IAVL
const commitInfoKeyFmt = "s/%d" // s/<version>

func GetCommitInfo(db dbm.DB, ver int64) (*types.CommitInfo, error) {
	cInfoKey := fmt.Sprintf(commitInfoKeyFmt, ver)

	bz, err := db.Get([]byte(cInfoKey))
	if err != nil {
		return nil, fmt.Errorf("failed to get commit info: %w", err)
	} else if bz == nil {
		return nil, errors.New("no commit info found")
	}

	cInfo := &types.CommitInfo{}
	if err = cInfo.Unmarshal(bz); err != nil {
		return nil, fmt.Errorf("failed unmarshal commit info: %w", err)
	}

	return cInfo, nil
}
