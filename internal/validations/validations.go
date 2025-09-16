package validations

import (
	registrycache "github.com/kyma-project/kim-snatch/api/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func ValidateConfig(config *registrycache.RegistryCacheConfig, secrets []v1.Secret) field.ErrorList {
	return nil
}

func ValidateConfigUpdate(newConfig, oldConfig *registrycache.RegistryCacheConfig, secrets []v1.Secret) field.ErrorList {
	return nil
}
