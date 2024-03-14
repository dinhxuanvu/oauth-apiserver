package openapiconfig

import (
	"github.com/MakeNowJust/heredoc"

	extensionsapiserver "k8s.io/apiextensions-apiserver/pkg/apiserver"
	apiserverendpointsopenapi "k8s.io/apiserver/pkg/endpoints/openapi"
	genericapiserver "k8s.io/apiserver/pkg/server"
	aggregatorscheme "k8s.io/kube-aggregator/pkg/apiserver/scheme"
	openapicommon "k8s.io/kube-openapi/pkg/common"
	"k8s.io/kube-openapi/pkg/spec3"
	"k8s.io/kube-openapi/pkg/validation/spec"
	"k8s.io/kubernetes/pkg/api/legacyscheme"

	openapigenerated "github.com/openshift/oauth-apiserver/pkg/openapi"
	"github.com/openshift/oauth-apiserver/pkg/version"

	// make sure our APIs are in the legacyscheme
	_ "github.com/openshift/oauth-apiserver/pkg/api/install"
)

var specInfo = spec.Info{
	InfoProps: spec.InfoProps{
		Title:   "OpenShift OAuth-related APIs",
		Version: version.Get().String(),
		License: &spec.License{
			Name: "Apache 2.0 (ASL2.0)",
			URL:  "http://www.apache.org/licenses/LICENSE-2.0",
		},
		Description: heredoc.Doc(`
			OpenShift OAuth APIs provide access and authorization tokens,
			users, groups and similar objects required for OpenShift integrated
			OAuth authentication to work on top of Kubernetes. The API allows
			consistent management of those objects.

			All API operations are authenticated via an Authorization bearer token that
			is provided for service accounts as a generated secret (in JWT form) or via
			the native OAuth access tokens. Core infrastructure components may use client
			certificates that require no authentication.

			All API operations return a 'resourceVersion' string that represents the
			version of the object in the underlying storage. The standard LIST operation
			performs a snapshot read of the underlying objects, returning a resourceVersion
			representing a consistent version of the listed objects. The WATCH operation
			allows all updates to a set of objects after the provided resourceVersion to
			be observed by a client. By listing and beginning a watch from the returned
			resourceVersion, clients may observe a consistent view of the state of one
			or more objects. Note that WATCH always returns the update after the provided
			resourceVersion. Watch may be extended a limited time in the past - using
			etcd 2 the watch window is 1000 events (which on a large cluster may only
			be a few tens of seconds) so clients must explicitly handle the "watch
			to old error" by re-listing.

			Objects are divided into two rough categories - those that have a lifecycle
			and must reflect the state of the cluster, and those that have no state.
			Objects with lifecycle typically have three main sections:

			* 'metadata' common to all objects
			* a 'spec' that represents the desired state
			* a 'status' that represents how much of the desired state is reflected on
				the cluster at the current time

			Objects that have no state have 'metadata' but may lack a 'spec' or 'status'
			section.

			Objects are divided into those that are namespace scoped (only exist inside
			of a namespace) and those that are cluster scoped (exist outside of
			a namespace). A namespace scoped resource will be deleted when the namespace
			is deleted and cannot be created if the namespace has not yet been created
			or is in the process of deletion. Cluster scoped resources are typically
			only accessible to admins - resources like nodes, persistent volumes, and
			cluster policy.

			All objects have a schema that is a combination of the 'kind' and
			'apiVersion' fields. This schema is additive only for any given version -
			no backwards incompatible changes are allowed without incrementing the
			apiVersion. The server will return and accept a number of standard
			responses that share a common schema - for instance, the common
			error type is 'metav1.Status' (described below) and will be returned
			on any error from the API server.

			The API is available in multiple serialization formats - the default is
			JSON (Accept: application/json and Content-Type: application/json) but
			clients may also use YAML (application/yaml) or the native Protobuf
			schema (application/vnd.kubernetes.protobuf). Note that the format
			of the WATCH API call is slightly different - for JSON it returns newline
			delimited objects while for Protobuf it returns length-delimited frames
			(4 bytes in network-order) that contain a 'versioned.Watch' Protobuf
			object.

			See the OpenShift documentation at https://docs.openshift.org for more
			information.
		`),
	},
}

func DefaultOpenAPIConfig() *openapicommon.Config {
	securityDefinitions := spec.SecurityDefinitions{}
	securityDefinitions["BearerToken"] = &spec.SecurityScheme{
		SecuritySchemeProps: spec.SecuritySchemeProps{
			Type:        "apiKey",
			Name:        "authorization",
			In:          "header",
			Description: "Bearer Token authentication",
		},
	}
	defNamer := apiserverendpointsopenapi.NewDefinitionNamer(legacyscheme.Scheme, extensionsapiserver.Scheme, aggregatorscheme.Scheme)
	return &openapicommon.Config{
		ProtocolList:      []string{"https"},
		GetDefinitions:    openapigenerated.GetOpenAPIDefinitions,
		IgnorePrefixes:    []string{"/swaggerapi", "/healthz", "/controllers", "/metrics", "/version/openshift", "/brokers"},
		GetDefinitionName: defNamer.GetDefinitionName,
		Info:              &specInfo,
		DefaultResponse: &spec.Response{
			ResponseProps: spec.ResponseProps{
				Description: "Default Response.",
			},
		},
		SecurityDefinitions: &securityDefinitions,
	}
}

func DefaultOpenAPIV3Config() *openapicommon.OpenAPIV3Config {
	defNamer := apiserverendpointsopenapi.NewDefinitionNamer(legacyscheme.Scheme, extensionsapiserver.Scheme, aggregatorscheme.Scheme)
	cfg := genericapiserver.DefaultOpenAPIV3Config(openapigenerated.GetOpenAPIDefinitions, defNamer)
	cfg.Info = &specInfo
	cfg.IgnorePrefixes = []string{"/swaggerapi", "/healthz", "/controllers", "/metrics", "/version/openshift", "/brokers"}
	cfg.SecuritySchemes = map[string]*spec3.SecurityScheme{
		"BearerToken": {
			SecuritySchemeProps: spec3.SecuritySchemeProps{
				Type:        "apiKey",
				Name:        "authorization",
				In:          "header",
				Description: "Bearer Token authentication",
			},
		},
	}

	return cfg
}
