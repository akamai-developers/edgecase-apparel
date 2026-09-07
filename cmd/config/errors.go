package config

import "errors"

var (
	errStkRefType    = errors.New("invalid type returned from stack reference")
	errStkRefNoValue = errors.New("no value returned from stack reference")
	errInvalidKey    = errors.New("key must be one of 'primary' or 'backup'")
	errKubeConfig    = errors.New("invalid kubeconfig")
	errGetProjRoot   = errors.New("repository root not found")
)
