package lpseal

import (
	"context"
	"github.com/filecoin-project/lotus/lib/harmony/harmonydb"
	"github.com/filecoin-project/lotus/lib/harmony/harmonytask"
	"github.com/filecoin-project/lotus/lib/promise"
	logging "github.com/ipfs/go-log/v2"
	"golang.org/x/xerrors"
	"time"
)

var log = logging.Logger("lpseal")

const (
	pollerSDR = iota

	numPollers
)

const sealPollerInterval = 10 * time.Second

type SealPoller struct {
	db *harmonydb.DB

	pollers [numPollers]promise.Promise[harmonytask.AddTaskFunc]
}

func (s *SealPoller) RunPoller(ctx context.Context) error {
	ticker := time.NewTicker(sealPollerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			s.poll(ctx)
		}
	}
}

func (s *SealPoller) poll(ctx context.Context) error {
	var tasks []struct {
		SpID         int64 `db:"sp_id"`
		SectorNumber int64 `db:"sector_number"`

		TaskSDR  *int64 `db:"task_id_sdr"`
		AfterSDR bool   `db:"after_sdr"`

		TaskTreeD  *int64 `db:"task_id_tree_d"`
		AfterTreeD bool   `db:"after_tree_d"`

		TaskTreeC  *int64 `db:"task_id_tree_c"`
		AfterTreeC bool   `db:"after_tree_c"`

		TaskTreeR  *int64 `db:"task_id_tree_r"`
		AfterTreeR bool   `db:"after_tree_r"`

		TaskPrecommitMsg  *int64 `db:"task_id_precommit_msg"`
		AfterPrecommitMsg bool   `db:"after_precommit_msg"`

		TaskPrecommitMsgWait     *int64 `db:"task_id_precommit_msg_wait"`
		AfterPrecommitMsgSuccess bool   `db:"after_precommit_msg_success"`

		TaskPoRep  *int64 `db:"task_id_porep"`
		PoRepProof []byte `db:"porep_proof"`

		TaskCommitMsg  *int64 `db:"task_id_commit_msg"`
		AfterCommitMsg bool   `db:"after_commit_msg"`

		TaskCommitMsgWait     *int64 `db:"task_id_commit_msg_wait"`
		AfterCommitMsgSuccess bool   `db:"after_commit_msg_success"`

		Failed       bool   `db:"failed"`
		FailedReason string `db:"failed_reason"`
	}

	err := s.db.Select(ctx, &tasks, `SELECT * FROM sectors_sdr_pipeline WHERE after_commit_msg_success != true`)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		if task.Failed {
			continue
		}

		if task.TaskSDR == nil {
			s.pollers[pollerSDR].Val(ctx)(func(id harmonytask.TaskID, tx *harmonydb.Tx) (shouldCommit bool, seriousError error) {
				n, err := tx.Exec(`UPDATE sectors_sdr_pipeline SET task_id_sdr = $1 WHERE sp_id = $2 AND sector_number = $3`, id, task.SpID, task.SectorNumber)
				if err != nil {
					return false, xerrors.Errorf("update sectors_sdr_pipeline: %w", err)
				}
				if n != 1 {
					return false, xerrors.Errorf("expected to update 1 row, updated %d", n)
				}

				return true, nil
			})
		}
		if task.TaskTreeD == nil {
			// todo start tree d task
		}

		// todo those two are really one pc2
		if task.TaskTreeC == nil && task.AfterSDR {
			// todo start tree c task
		}
		if task.TaskTreeR == nil && task.AfterTreeC {
			// todo start tree r task
		}

		if task.TaskPrecommitMsg == nil && task.AfterTreeR && task.AfterTreeD {
			// todo start precommit msg task
		}

		if task.TaskPrecommitMsgWait == nil && task.AfterPrecommitMsg {
			// todo start precommit msg wait task
		}

		todoWaitSeed := false
		if task.TaskPoRep == nil && todoWaitSeed {
			// todo start porep task
		}
	}

	return nil
}
