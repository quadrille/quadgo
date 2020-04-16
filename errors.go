package quadgo

import "errors"

var (
	ErrNoLeaderFound  = errors.New("no leader found")
	ErrEmptyBulkWrite = errors.New("cannot execute empty bulkwrite operation")
)
