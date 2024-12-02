package config

import "github.com/lucheng0127/bumblebee/pkg/utils/validation"

func (cfg *ApiServerConfig) validatePlatformConfig() []error {
	var errs []error

	if err := validation.ValidatePort(cfg.Platform.Port); err != nil {
		errs = append(errs, err)
	}
	if err := validation.ValidateName(cfg.Platform.Zone); err != nil {
		errs = append(errs, err)
	}

	return errs
}

func (cfg *ApiServerConfig) Validate() []error {
	var errs []error

	errs = append(errs, cfg.validatePlatformConfig()...)

	return errs
}
