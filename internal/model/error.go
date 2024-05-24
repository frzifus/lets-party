// Copyright (C) 2024 the quixsi maintainers
// See root-dir/LICENSE for more information

package model

type ErrorReason int

const (
	ErrorReasonDeadline ErrorReason = iota
	ErrorReasonProcess
)
