package validations

import (
	registrycache "github.com/kyma-project/kim-snatch/api/v1beta1"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/ptr"
	"strings"
	"testing"
)

const (
	InvalidUpstreamPort           = "docker.io:77777"
	InvalidVolumeSize             = "-1"
	InvalidVolumeStorageClassName = "Invalid.Name"
	InvalidGarbageCollectionValue = -1
	InvalidHttpProxyUrl           = "http//invalid-url"
	InvalidHttpsProxyUrl          = "https//invalid-url"
)

func TestDo(t *testing.T) {

	upstreamFieldPath := field.NewPath("spec").Child("upstream")
	volumeSizeFieldPath := field.NewPath("spec").Child("volume").Child("size")
	volumeStorageClassNameFieldPath := field.NewPath("spec").Child("volume").Child("storageClassName")
	garbageCollectionTTLFieldPath := field.NewPath("spec").Child("garbageCollection").Child("ttl")
	httpProxyFieldPath := field.NewPath("spec").Child("volume").Child("proxy").Child("httpProxy")
	httpsProxyFieldPath := field.NewPath("spec").Child("volume").Child("proxy").Child("httpsProxy")

	secretWithIncorrectStructure := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "invalid-secret",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"invalid-key": []byte("invalid-value"),
		},
		Immutable: ptr.To(false),
	}

	for _, tt := range []struct {
		name string
		registrycache.RegistryCacheConfig
		errorsList field.ErrorList
		secrets    []v1.Secret
	}{
		{
			name: "empty spec",
			RegistryCacheConfig: registrycache.RegistryCacheConfig{
				Spec: registrycache.RegistryCacheConfigSpec{},
			},
			errorsList: field.ErrorList{
				field.Required(field.NewPath("spec"), "spec cannot be empty"),
			},
		},
		{
			name: "valid spec",
			RegistryCacheConfig: registrycache.RegistryCacheConfig{
				Spec: registrycache.RegistryCacheConfigSpec{
					Upstream: "docker.io",
				},
			},
			errorsList: field.ErrorList{},
		},
		{
			name: "invalid spec",
			RegistryCacheConfig: registrycache.RegistryCacheConfig{
				Spec: registrycache.RegistryCacheConfigSpec{
					Upstream: InvalidUpstreamPort,
					Volume: &registrycache.Volume{
						Size:             ptr.To(resource.MustParse(InvalidVolumeSize)),
						StorageClassName: ptr.To(InvalidVolumeStorageClassName),
					},
					GarbageCollection: &registrycache.GarbageCollection{
						TTL: metav1.Duration{Duration: InvalidGarbageCollectionValue},
					},
					Proxy: &registrycache.Proxy{
						HTTPProxy:  ptr.To(InvalidHttpProxyUrl),
						HTTPSProxy: ptr.To(InvalidHttpsProxyUrl),
					},
				},
			},
			errorsList: field.ErrorList{
				field.Invalid(upstreamFieldPath, InvalidUpstreamPort, "valid port must be in the range [1, 65535]"),
				field.Invalid(volumeSizeFieldPath, InvalidVolumeSize, "must be greater than 0"),
				field.Invalid(volumeStorageClassNameFieldPath, InvalidVolumeStorageClassName, "an RFC 1123 subdomain must consist of alphanumeric characters"),
				field.Invalid(garbageCollectionTTLFieldPath, InvalidGarbageCollectionValue, "ttl must be a non-negative duration"),
				field.Invalid(httpProxyFieldPath, InvalidHttpProxyUrl, "some error"),
				field.Invalid(httpsProxyFieldPath, InvalidHttpsProxyUrl, "some error"),
			},
		},
		{
			name: "non existent secret reference name",
			RegistryCacheConfig: registrycache.RegistryCacheConfig{
				Spec: registrycache.RegistryCacheConfigSpec{
					Upstream:            "docker.io",
					SecretReferenceName: ptr.To("non-existent-secret"),
				},
			},
			errorsList: field.ErrorList{
				field.NotFound(field.NewPath("spec").Child("secretReferenceName"), "non-existent-secret"),
			},
		},
		{
			name: "secret with incorrect structure",
			secrets: []v1.Secret{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "invalid-secret",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"invalid-key": []byte("invalid-value"),
					},
					Immutable: ptr.To(true),
				},
			},
			RegistryCacheConfig: registrycache.RegistryCacheConfig{
				Spec: registrycache.RegistryCacheConfigSpec{
					Upstream:            "docker.io",
					SecretReferenceName: ptr.To("invalid-secret"),
				},
			},
			errorsList: field.ErrorList{
				field.NotFound(field.NewPath("spec").Child("secretReferenceName"), "invalid-secret"),
			},
		},
		{
			name: "mutable secret",
			secrets: []v1.Secret{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "invalid-secret",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"invalid-key": []byte("invalid-value"),
					},
					Immutable: ptr.To(false),
				},
			},
			RegistryCacheConfig: registrycache.RegistryCacheConfig{
				Spec: registrycache.RegistryCacheConfigSpec{
					Upstream:            "docker.io",
					SecretReferenceName: ptr.To("invalid-secret"),
				},
			},
			errorsList: field.ErrorList{
				field.Invalid(field.NewPath("spec").Child("secretReferenceName"), secretWithIncorrectStructure, "should be immutable"),
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateConfig(&tt.RegistryCacheConfig, tt.secrets)

			require.Equal(t, len(tt.errorsList), len(errs))

			for _, expectedErr := range tt.errorsList {
				var actualFieldError *field.Error

				for _, actualErr := range errs {
					if actualErr.Type == expectedErr.Type && expectedErr.Field == actualErr.Field {
						actualFieldError = actualErr
						break
					}
				}
				require.NotNil(t, actualFieldError, "expected error not found: %v", expectedErr)

				require.Equal(t, expectedErr.BadValue, actualFieldError.BadValue)
				require.True(t, strings.Contains(actualFieldError.Detail, expectedErr.Detail))
			}
		})
	}
}
