/*
Copyright 2016 Skippbox, Ltd.

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

package controller

import (
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/skippbox/kubewatch/config"
	"github.com/skippbox/kubewatch/pkg/handlers"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/cache"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/controller/framework"
	"k8s.io/kubernetes/pkg/fields"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/util/wait"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/watch"
	"k8s.io/kubernetes/pkg/apis/extensions"
)

func Controller(conf *config.Config, eventHandler handlers.Handler) {

	factory := cmdutil.NewFactory(nil)
	kubeConfig, err := factory.ClientConfig()
	if err != nil {
		logrus.Fatal(err)
	}

	kubeClient := client.NewOrDie(kubeConfig)

	if conf.Resource.Pod {
		var podsStore cache.Store
		podsStore = watchPods(kubeClient, podsStore, eventHandler)
	}

	if conf.Resource.Services {
		var servicesStore cache.Store
		servicesStore = watchServices(kubeClient, servicesStore, eventHandler)
	}

	if conf.Resource.ReplicationController {
		var rcStore cache.Store
		rcStore = watchReplicationControllers(kubeClient, rcStore, eventHandler)
	}

	if conf.Resource.Deployment {
		var depStore cache.Store
		depStore = watchDeployments(kubeClient, depStore, eventHandler)
	}

	if conf.Resource.ReplicaSet {
		var tprStore cache.Store
		tprStore = watchThirdPartyResources(kubeClient, tprStore, eventHandler)
	}

	logrus.Fatal(http.ListenAndServe(":8081", nil))
}

func watchDeployments(client *client.Client, store cache.Store, eventHandler handlers.Handler) cache.Store {
	resyncPeriod := 30 * time.Minute
	eStore, eController := framework.NewIndexerInformer(
		&cache.ListWatch{
			ListFunc: func(options api.ListOptions) (runtime.Object, error) {
				return client.Extensions().Deployments(api.NamespaceAll).List(options)
			},
			WatchFunc: func(options api.ListOptions) (watch.Interface, error) {
				return client.Extensions().Deployments(api.NamespaceAll).Watch(options)
			},
		},
		&extensions.Deployment{},
		resyncPeriod,
		framework.ResourceEventHandlerFuncs{
			AddFunc:    eventHandler.ObjectCreated,
			UpdateFunc: eventHandler.ObjectUpdated,
			DeleteFunc: eventHandler.ObjectDeleted,
		},
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)

	go eController.Run(wait.NeverStop)

	return eStore
}

func watchThirdPartyResources(client *client.Client, store cache.Store, eventHandler handlers.Handler) cache.Store {
	resyncPeriod := 30 * time.Minute
	eStore, eController := framework.NewIndexerInformer(
		&cache.ListWatch{
			ListFunc: func(options api.ListOptions) (runtime.Object, error) {
				return client.Extensions().ThirdPartyResources().List(options)
			},
			WatchFunc: func(options api.ListOptions) (watch.Interface, error) {
				return client.Extensions().ThirdPartyResources().Watch(options)
			},
		},
		&extensions.ThirdPartyResource{},
		resyncPeriod,
		framework.ResourceEventHandlerFuncs{
			AddFunc:    eventHandler.ObjectCreated,
			UpdateFunc: eventHandler.ObjectUpdated,
			DeleteFunc: eventHandler.ObjectDeleted,
		},
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)

	go eController.Run(wait.NeverStop)

	return eStore
}

func watchPods(client *client.Client, store cache.Store, eventHandler handlers.Handler) cache.Store {
	//Define what we want to look for (Pods)
	watchlist := cache.NewListWatchFromClient(client, "pods", api.NamespaceAll, fields.Everything())

	resyncPeriod := 30 * time.Minute

	//Setup an informer to call functions when the watchlist changes
	eStore, eController := framework.NewInformer(
		watchlist,
		&api.Pod{},
		resyncPeriod,
		framework.ResourceEventHandlerFuncs{
			AddFunc:    eventHandler.ObjectCreated,
			DeleteFunc: eventHandler.ObjectDeleted,
			UpdateFunc: eventHandler.ObjectUpdated,
		},
	)

	//Run the controller as a goroutine
	go eController.Run(wait.NeverStop)

	return eStore
}

func watchServices(client *client.Client, store cache.Store, eventHandler handlers.Handler) cache.Store {
	//Define what we want to look for (Services)
	watchlist := cache.NewListWatchFromClient(client, "services", api.NamespaceAll, fields.Everything())

	resyncPeriod := 30 * time.Minute

	//Setup an informer to call functions when the watchlist changes
	eStore, eController := framework.NewInformer(
		watchlist,
		&api.Service{},
		resyncPeriod,
		framework.ResourceEventHandlerFuncs{
			AddFunc:    eventHandler.ObjectCreated,
			DeleteFunc: eventHandler.ObjectDeleted,
			UpdateFunc: eventHandler.ObjectUpdated,
		},
	)

	//Run the controller as a goroutine
	go eController.Run(wait.NeverStop)

	return eStore
}

func watchReplicationControllers(client *client.Client, store cache.Store, eventHandler handlers.Handler) cache.Store {
	//Define what we want to look for (ReplicationControllers)
	watchlist := cache.NewListWatchFromClient(client, "replicationcontrollers", api.NamespaceAll, fields.Everything())

	resyncPeriod := 30 * time.Minute

	//Setup an informer to call functions when the watchlist changes
	eStore, eController := framework.NewInformer(
		watchlist,
		&api.ReplicationController{},
		resyncPeriod,
		framework.ResourceEventHandlerFuncs{
			AddFunc:    eventHandler.ObjectCreated,
			DeleteFunc: eventHandler.ObjectDeleted,
			UpdateFunc: eventHandler.ObjectUpdated,
		},
	)

	//Run the controller as a goroutine
	go eController.Run(wait.NeverStop)

	return eStore
}
