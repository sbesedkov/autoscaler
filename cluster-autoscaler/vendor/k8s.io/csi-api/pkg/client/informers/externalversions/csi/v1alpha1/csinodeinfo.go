/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	time "time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
	csiv1alpha1 "k8s.io/csi-api/pkg/apis/csi/v1alpha1"
	versioned "k8s.io/csi-api/pkg/client/clientset/versioned"
	internalinterfaces "k8s.io/csi-api/pkg/client/informers/externalversions/internalinterfaces"
	v1alpha1 "k8s.io/csi-api/pkg/client/listers/csi/v1alpha1"
)

// CSINodeInfoInformer provides access to a shared informer and lister for
// CSINodeInfos.
type CSINodeInfoInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.CSINodeInfoLister
}

type cSINodeInfoInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// NewCSINodeInfoInformer constructs a new informer for CSINodeInfo type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewCSINodeInfoInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredCSINodeInfoInformer(client, resyncPeriod, indexers, nil)
}

// NewFilteredCSINodeInfoInformer constructs a new informer for CSINodeInfo type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredCSINodeInfoInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.CsiV1alpha1().CSINodeInfos().List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.CsiV1alpha1().CSINodeInfos().Watch(options)
			},
		},
		&csiv1alpha1.CSINodeInfo{},
		resyncPeriod,
		indexers,
	)
}

func (f *cSINodeInfoInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredCSINodeInfoInformer(client, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *cSINodeInfoInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&csiv1alpha1.CSINodeInfo{}, f.defaultInformer)
}

func (f *cSINodeInfoInformer) Lister() v1alpha1.CSINodeInfoLister {
	return v1alpha1.NewCSINodeInfoLister(f.Informer().GetIndexer())
}