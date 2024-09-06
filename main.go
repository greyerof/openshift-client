package main

import (
	"context"
	"fmt"
	"os"

	apiserverv1 "github.com/openshift/api/apiserver/v1"
	configv1 "github.com/openshift/api/config/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var scheme *runtime.Scheme

func init() {
	scheme = runtime.NewScheme()
}

func getClient() (client.WithWatch, error) {
	// Under the hood, all k8s native types should've been registered like:
	// 	SchemeBuilder.Register(&CertsuiteRun{}, &CertsuiteRunList{})
	//  That call will register new Object types that must implement DeepCopy & Kind:
	// type Object interface {
	// 	GetObjectKind() schema.ObjectKind
	// 	DeepCopyObject() Object
	// }

	err := clientgoscheme.AddToScheme(scheme)
	if err != nil {
		return nil, fmt.Errorf("failed to install k8s schemes: %v", err)
	}

	// This also works, but the installed version is not known. Probably the last official one... (?)
	//   err = apiserver.Install(scheme)
	// Needs this: import "github.com/openshift/api/apiserver"
	err = apiserverv1.AddToScheme(scheme)
	if err != nil {
		return nil, fmt.Errorf("failed to install apiserver scheme: %v", err)
	}

	restConfig, err := ctrl.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get rest config: %v", err)
	}

	// k8sClient, err := client.New(restConfig, client.Options{})
	// Interesting fact: both client.New() and client.NewWithWatch return interfaces, not structs,
	// which contradicts the go proverb "accept interfaces, return structs".
	k8sClient, err := client.NewWithWatch(restConfig, client.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s client: %v", err)
	}

	return k8sClient, nil
}

func main() {
	var c client.WithWatch
	c, err := getClient()
	if err != nil {
		fmt.Printf("Failed to get client: %v\n", err)
		os.Exit(1)
	}

	// Let's get the ApiRequestCount
	apireqslist := apiserverv1.APIRequestCountList{}
	err = c.List(context.TODO(), &apireqslist)
	if err != nil {
		fmt.Printf("Failed to get apirequestcounts: %v\n", err)
		os.Exit(1)
	}

	if len(apireqslist.Items) == 0 {
		fmt.Println("No ApiRequestCounts found.")
	} else {
		for i := range apireqslist.Items {
			apireq := &apireqslist.Items[i]
			fmt.Println(apireq.Name)
		}
	}

	// Let's add a new type on the fly and see if it works
	fmt.Println("Getting cluster consoles. Its types are added to the scheme after client creation.")
	configv1.AddToScheme(scheme)
	if err != nil {
		fmt.Printf("Failed to get install console types: %v\n", err)
		os.Exit(1)
	}

	consoleList := configv1.ConsoleList{}
	err = c.List(context.TODO(), &consoleList)
	if err != nil {
		fmt.Printf("Failed to get console list: %v\n", err)
		os.Exit(1)
	}

	if len(consoleList.Items) == 0 {
		fmt.Println("No Consoles found.")
	} else {
		for i := range consoleList.Items {
			apireq := &consoleList.Items[i]
			fmt.Println(apireq.Name)
		}
	}
}
