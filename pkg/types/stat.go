package types

import (
	"context"

	"github.com/rs/xstats"
)

// StatsProvider extracts a stat client from context.
type StatsProvider func(ctx context.Context) xstats.XStater
