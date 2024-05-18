// Copyright (C) 2024 the lets-party maintainers
// See root-dir/LICENSE for more information

package model

type ErrorReason int

const (
	ErrorReasonDeadline ErrorReason = iota
	ErrorReasonProcess
)
