//go:build !linux

package execution

import "context"

func acquireHomebrewLinux(context.Context) HomebrewAcquisitionResult {
	return HomebrewAcquisitionResult{Err: ErrHomebrewAcquisitionUnavailable}
}
