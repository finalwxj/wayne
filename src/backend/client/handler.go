// Proxy is a package responsible for doing common operations on kubernetes resources
// like UPDATE DELETE CREATE GET deployment and so on.
package client

import (
	"fmt"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"

	"github.com/Qihoo360/wayne/src/backend/client/api"
	"github.com/Qihoo360/wayne/src/backend/util/logs"
)

type ResourceHandler interface {
	Put(kind string, namespace string, name string, object *runtime.Unknown) error
	Get(kind string, namespace string, name string) (runtime.Object, error)
	List(kind string, namespace string, labelSelector string) ([]runtime.Object, error)
	Delete(kind string, namespace string, name string, options *meta_v1.DeleteOptions) error
}

type resourceHandler struct {
	client       *kubernetes.Clientset
	cacheFactory *CacheFactory
}

func NewResourceHandler(kubeClient *kubernetes.Clientset, cacheFactory *CacheFactory) ResourceHandler {
	return &resourceHandler{
		client:       kubeClient,
		cacheFactory: cacheFactory,
	}
}

func (h *resourceHandler) Put(kind string, namespace string, name string, object *runtime.Unknown) error {
	resource, ok := api.KindToResourceMap[kind]
	if !ok {
		return fmt.Errorf("Resource kind (%s) not support yet . ", kind)
	}
	req := h.client.RESTClient().Put().
		Resource(kind).
		SetHeader("Content-Type", "application/json").
		Body([]byte(object.Raw))
	if resource.Namespaced {
		req.Namespace(namespace)
	}
	return req.Do().Error()
}

func (h *resourceHandler) Delete(kind string, namespace string, name string, options *meta_v1.DeleteOptions) error {
	resource, ok := api.KindToResourceMap[kind]
	if !ok {
		return fmt.Errorf("Resource kind (%s) not support yet . ", kind)
	}

	req := h.client.RESTClient().Put().
		Resource(kind).
		Name(name).
		Body(options)
	if resource.Namespaced {
		req.Namespace(namespace)
	}

	return req.Do().Error()
}

// Get object from cache
func (h *resourceHandler) Get(kind string, namespace string, name string) (runtime.Object, error) {
	resource, ok := api.KindToResourceMap[kind]
	if !ok {
		return nil, fmt.Errorf("Resource kind (%s) not support yet . ", kind)
	}
	genericInformer, err := h.cacheFactory.sharedInformerFactory.ForResource(resource.GroupVersionResource)
	if err != nil {
		return nil, err
	}
	lister := genericInformer.Lister()
	if resource.Namespaced {
		lister.ByNamespace(namespace)
	}

	return lister.Get(name)
}

// Get object from cache
func (h *resourceHandler) List(kind string, namespace string, labelSelector string) ([]runtime.Object, error) {
	resource, ok := api.KindToResourceMap[kind]
	if !ok {
		return nil, fmt.Errorf("Resource kind (%s) not support yet . ", kind)
	}
	genericInformer, err := h.cacheFactory.sharedInformerFactory.ForResource(resource.GroupVersionResource)
	if err != nil {
		return nil, err
	}
	lister := genericInformer.Lister()
	if resource.Namespaced {
		lister.ByNamespace(namespace)
	}

	selectors, err := labels.Parse(labelSelector)
	if err != nil {
		logs.Error("Build label selector error.", err)
		return nil, err
	}

	return lister.List(selectors)
}
