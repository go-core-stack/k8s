package apply

import (
	"context"
	"fmt"
	"log"

	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	uns "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ApplyObject applies the desired object against the apiserver,
// merging it with any existing objects if already present.
func ApplyObject(ctx context.Context, client k8sclient.Client, obj *uns.Unstructured) error {
	name := obj.GetName()
	namespace := obj.GetNamespace()
	if name == "" {
		return errors.Errorf("Object %s has no name", obj.GroupVersionKind().String())
	}
	gvk := obj.GroupVersionKind()
	// used for logging and errors
	objDesc := fmt.Sprintf("(%s) %s/%s", gvk.String(), namespace, name)
	log.Printf("reconciling %s", objDesc)

	if err := IsObjectSupported(obj); err != nil {
		return errors.Wrapf(err, "object %s unsupported", objDesc)
	}

	// Get existing
	existing := &uns.Unstructured{}
	existing.SetGroupVersionKind(gvk)
	err := client.Get(ctx, types.NamespacedName{Name: obj.GetName(), Namespace: obj.GetNamespace()}, existing)

	if err != nil && apierrors.IsNotFound(err) {
		log.Printf("does not exist, creating %s", objDesc)
		err := client.Create(ctx, obj)
		if err != nil {
			return errors.Wrapf(err, "could not create %s", objDesc)
		}
		log.Printf("successfully created %s", objDesc)
		return nil
	}
	if err != nil {
		return errors.Wrapf(err, "could not retrieve existing %s", objDesc)
	}

	// Merge the desired object with what actually exists
	if err := MergeObjectForUpdate(existing, obj); err != nil {
		return errors.Wrapf(err, "could not merge object %s with existing", objDesc)
	}
	if !equality.Semantic.DeepEqual(existing, obj) {
		if err := client.Update(ctx, obj); err != nil {
			return errors.Wrapf(err, "could not update object %s", objDesc)
		} else {
			log.Printf("update was successful")
		}
	}

	return nil
}

// ApplyObjectServerSide applies the desired object using server-side apply
// (https://kubernetes.io/docs/reference/using-api/server-side-apply/) under the
// given field manager, instead of the read-modify-write Update that ApplyObject
// performs. Because SSA does not use a resourceVersion, it never conflicts with
// another controller writing a different part of the same object (e.g. an
// operator owning .status while the caller owns .spec). It is idempotent and
// creates the object if absent.
//
// fieldManager must be a stable, caller-unique name; ownership of the fields set
// here is recorded under it. ForceOwnership is used so the caller takes over any
// fields previously written by a client-side-apply / Update manager (e.g. when
// migrating an object from ApplyObject to this function).
//
// NOTE: SSA prunes fields this manager previously set but no longer sets, so the
// caller must render the COMPLETE set of fields it owns on every call.
func ApplyObjectServerSide(ctx context.Context, client k8sclient.Client, obj *uns.Unstructured, fieldManager string) error {
	name := obj.GetName()
	namespace := obj.GetNamespace()
	if name == "" {
		return errors.Errorf("Object %s has no name", obj.GroupVersionKind().String())
	}
	if fieldManager == "" {
		return errors.Errorf("Object (%s) %s/%s: fieldManager is required for server-side apply",
			obj.GroupVersionKind().String(), namespace, name)
	}
	objDesc := fmt.Sprintf("(%s) %s/%s", obj.GroupVersionKind().String(), namespace, name)
	log.Printf("server-side applying %s (fieldManager=%s)", objDesc, fieldManager)

	if err := IsObjectSupported(obj); err != nil {
		return errors.Wrapf(err, "object %s unsupported", objDesc)
	}

	// SSA must not carry a resourceVersion (it does not use optimistic
	// concurrency) nor managedFields (the apiserver maintains those).
	obj.SetResourceVersion("")
	obj.SetManagedFields(nil)

	if err := client.Patch(ctx, obj, k8sclient.Apply,
		k8sclient.FieldOwner(fieldManager), k8sclient.ForceOwnership); err != nil {
		return errors.Wrapf(err, "could not server-side apply %s", objDesc)
	}
	return nil
}
