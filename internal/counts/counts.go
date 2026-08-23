package counts

import "sync/atomic"

var WriteMono, FillBuffer, DisplayTick, ScanRow atomic.Uint32
