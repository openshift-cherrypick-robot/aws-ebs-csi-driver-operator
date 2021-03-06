package operator

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/openshift/aws-ebs-csi-driver-operator/pkg/generated"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
)

type fakeDynamicClient struct {
	credentialRequest *unstructured.Unstructured
	t                 *testing.T
}

var (
	_ dynamic.Interface                      = &fakeDynamicClient{}
	_ dynamic.ResourceInterface              = &fakeDynamicClient{}
	_ dynamic.NamespaceableResourceInterface = &fakeDynamicClient{}

	credentialsGR schema.GroupResource = schema.GroupResource{Group: credentialsRequestGroup, Resource: credentialsRequestResource}
)

func (fake *fakeDynamicClient) Resource(resource schema.GroupVersionResource) dynamic.NamespaceableResourceInterface {
	return fake
}

func (fake *fakeDynamicClient) Namespace(string) dynamic.ResourceInterface {
	return fake
}

func (fake *fakeDynamicClient) Create(ctx context.Context, obj *unstructured.Unstructured, options metav1.CreateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	if err := checkCredentialsRequestSanity(obj); err != nil {
		return nil, err
	}
	if fake.credentialRequest != nil {
		return nil, apierrors.NewAlreadyExists(credentialsGR, obj.GetName())
	}
	fake.credentialRequest = obj.DeepCopy()
	fake.credentialRequest.SetGeneration(1)
	fake.credentialRequest.SetResourceVersion("1")
	return fake.credentialRequest, nil
}

func (fake *fakeDynamicClient) Update(ctx context.Context, obj *unstructured.Unstructured, options metav1.UpdateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	if err := checkCredentialsRequestSanity(obj); err != nil {
		return nil, err
	}
	if fake.credentialRequest == nil {
		return nil, apierrors.NewNotFound(credentialsGR, obj.GetName())
	}
	fake.credentialRequest = obj.DeepCopy()
	fake.credentialRequest.SetGeneration(obj.GetGeneration() + 1)
	gen, _ := strconv.Atoi(obj.GetResourceVersion())
	fake.credentialRequest.SetResourceVersion(strconv.Itoa(gen + 1))
	return fake.credentialRequest, nil
}

func (fake *fakeDynamicClient) UpdateStatus(ctx context.Context, obj *unstructured.Unstructured, options metav1.UpdateOptions) (*unstructured.Unstructured, error) {
	return nil, errors.New("not implemented")
}

func (fake *fakeDynamicClient) Delete(ctx context.Context, name string, options metav1.DeleteOptions, subresources ...string) error {
	if fake.credentialRequest == nil {
		return apierrors.NewNotFound(credentialsGR, name)
	}
	fake.credentialRequest = nil
	return nil
}

func (fake *fakeDynamicClient) DeleteCollection(ctx context.Context, options metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	return errors.New("not implemented")
}

func (fake *fakeDynamicClient) Get(ctx context.Context, name string, options metav1.GetOptions, subresources ...string) (*unstructured.Unstructured, error) {
	if fake.credentialRequest == nil {
		return nil, apierrors.NewNotFound(credentialsGR, name)
	}
	if fake.credentialRequest.GetName() != name {
		return nil, apierrors.NewNotFound(credentialsGR, name)
	}
	return fake.credentialRequest, nil
}

func (fake *fakeDynamicClient) List(ctx context.Context, opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	return nil, errors.New("not implemented")
}

func (fake *fakeDynamicClient) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return nil, errors.New("not implemented")
}

func (fake *fakeDynamicClient) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, options metav1.PatchOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, errors.New("not implemented")
}

func checkCredentialsRequestSanity(obj *unstructured.Unstructured) error {
	if obj.GetKind() != credentialsRequestKind {
		return fmt.Errorf("expected kind %s, got %s", credentialsRequestKind, obj.GetKind())
	}
	if obj.GetNamespace() != credentialRequestNamespace {
		return fmt.Errorf("expected namespace %s, got %s", credentialRequestNamespace, obj.GetNamespace())
	}
	if obj.GetName() != operandNamespace {
		return fmt.Errorf("expected namespace %s, got %s", operandNamespace, obj.GetNamespace())
	}

	expectedObj := readCredentialRequestsOrDie(generated.MustAsset(credentialsRequest))
	expectedSpec := expectedObj.Object["spec"]
	actualSpec := obj.Object["spec"]
	if !reflect.DeepEqual(expectedSpec, actualSpec) {
		// TODO: add diff
		return fmt.Errorf("expected different spec")
	}
	return nil
}
