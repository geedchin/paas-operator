package main

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"

	blueprintv1alpha1 "github.com/farmer-hutao/k6s/pkg/controller/apis/blueprintcontroller/v1alpha1"
	clientset "github.com/farmer-hutao/k6s/pkg/controller/client/clientset/versioned"
	blueprintscheme "github.com/farmer-hutao/k6s/pkg/controller/client/clientset/versioned/scheme"
	informers "github.com/farmer-hutao/k6s/pkg/controller/client/informers/externalversions/blueprintcontroller/v1alpha1"
	listers "github.com/farmer-hutao/k6s/pkg/controller/client/listers/blueprintcontroller/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

const controllerAgentName = "blueprint-controller"

const (
	// SuccessSynced is used as part of the Event 'reason' when a Database is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a Database fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by Database"
	// MessageResourceSynced is the message used for an Event fired when a Database
	// is synced successfully
	MessageResourceSynced = "Database synced successfully"
)

// Controller is the controller implementation for Database resources
type Controller struct {
	// blueprintclientset is a clientset for our own API group
	blueprintclientset clientset.Interface

	databasesLister listers.DatabaseLister
	databasesSynced cache.InformerSynced

	workqueue workqueue.RateLimitingInterface

	recorder record.EventRecorder
}

// NewController returns a new blueprint controller
func NewController(
	blueprintclientset clientset.Interface,
	databaseInformer informers.DatabaseInformer) *Controller {

	// Create event broadcaster
	// Add blueprint-controller types to the default Kubernetes Scheme so Events can be
	// logged for blueprint-controller types.
	utilruntime.Must(blueprintscheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	//eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		blueprintclientset: blueprintclientset,
		databasesLister:    databaseInformer.Lister(),
		databasesSynced:    databaseInformer.Informer().HasSynced,
		workqueue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Databases"),
		recorder:           recorder,
	}

	klog.Info("Setting up event handlers")
	// Set up an event handler for when Database resources change
	databaseInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueDatabase,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueDatabase(new)
		},
	})

	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting Database controller")

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.databasesSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	// Launch two workers to process Database resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// Database resource to be synced.
		if err := c.syncHandler(key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workqueue.AddRateLimited(key)
			fmt.Println("Err: ", key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		klog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Database resource
// with the current status of the resource.
func (c *Controller) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the Database resource with this namespace/name
	database, err := c.databasesLister.Databases(namespace).Get(name)
	if err != nil {
		// The Database resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("database '%s' in work queue no longer exists", key))
			return nil
		}

		return err
	}

	ip := database.Spec.Host.Ip
	fmt.Println("Ip: ", ip)

	// Get the deployment with the name specified in Database.spec

	// If the Deployment is not controlled by this Database resource, we should log
	// a warning to the event recorder and ret

	// Finally, we update the status block of the Database resource to reflect the
	// current state of the world
	err = c.updateDatabaseStatus(database, "not-installed")
	if err != nil {
		return err
	}

	c.recorder.Event(database, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func (c *Controller) updateDatabaseStatus(database *blueprintv1alpha1.Database, install string) error {
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	databaseCopy := database.DeepCopy()
	databaseCopy.Status.Install = install
	_, err := c.blueprintclientset.BlueprintcontrollerV1alpha1().Databases(database.Namespace).Update(databaseCopy)
	return err
}

// enqueueDatabase takes a Database resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than Database.
func (c *Controller) enqueueDatabase(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)

	fmt.Println("enqueue: ", key)
}
