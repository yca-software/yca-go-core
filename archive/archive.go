package yca_archive

import (
	"errors"
	"fmt"
	"time"

	yca_error "github.com/yca-software/yca-go-core/error"
)

type ArchiveFilter string

const (
	ArchiveFilterActive   ArchiveFilter = "active"
	ArchiveFilterArchived ArchiveFilter = "archived"
)

// DefaultArchiveFilter is used when no filter is specified.
const DefaultArchiveFilter = ArchiveFilterActive

// ArchivedDataRetentionPeriod is how long soft-deleted rows and stale/expired rows
// (tokens, invitations, API keys, etc.) are kept before background cleanup hard-deletes them.
const ArchivedDataRetentionPeriod = 30 * 24 * time.Hour

func NormalizeArchiveFilter(filter ArchiveFilter) (ArchiveFilter, error) {
	if filter == "" {
		return DefaultArchiveFilter, nil
	}
	if filter != ArchiveFilterActive && filter != ArchiveFilterArchived {
		return "", fmt.Errorf("invalid archive filter %q: must be active or archived", filter)
	}
	return filter, nil
}

func ParseArchiveFilterQuery(value string) (ArchiveFilter, error) {
	if value == "" {
		return DefaultArchiveFilter, nil
	}
	filter := ArchiveFilter(value)
	if filter != ArchiveFilterActive && filter != ArchiveFilterArchived {
		return "", yca_error.NewBadRequestError(errors.New("invalid archive filter"), "InvalidArchiveFilter", nil)
	}
	return filter, nil
}
