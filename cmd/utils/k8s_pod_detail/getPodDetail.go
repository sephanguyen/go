package k8spoddetail

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	namespaces []string
	podData    [][]string
	kubeConfig *string
	imageName  string
	tag        string
)

// parse a string data to a table row
func parseStringToRow(st []string) table.Row {
	var row table.Row
	for _, val := range st {
		row = append(row, val)
	}
	return row
}

// print information as a table
func printInfo(header []string, data [][]string) {
	rowConfigAutoMerge := table.RowConfig{AutoMerge: true}
	t := table.NewWriter()
	headerRow := parseStringToRow(header)
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(headerRow)
	for _, val := range data {
		t.AppendRow(parseStringToRow(val), rowConfigAutoMerge)
	}
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, AutoMerge: true},
		{Number: 2, AutoMerge: true},
		{Number: 3, AutoMerge: true},
		{Number: 4, AutoMerge: true},
		{Number: 5, Align: text.AlignCenter, AlignFooter: text.AlignCenter, AlignHeader: text.AlignCenter},
		{Number: 6, Align: text.AlignCenter, AlignFooter: text.AlignCenter, AlignHeader: text.AlignCenter},
	})
	t.SetAutoIndex(true)
	t.Style().Options.SeparateRows = true
	t.Render()
}

func getPodDetail(cmd *cobra.Command, args []string) error {
	if home := homedir.HomeDir(); home != "" {
		kubeConfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeConfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeConfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// if point a specific namespace then set the namespaces to one element: namespace
	if namespace != "" {
		namespaces = []string{namespace}
	} else {
		// get list all namespaces
		nameSpaceList, _ := clientSet.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
		fmt.Printf("There are %d namespace(s) in the cluster: \n", len(nameSpaceList.Items))
		for _, val := range nameSpaceList.Items {
			fmt.Printf("%v \n", val.Name)
			namespaces = append(namespaces, val.Name)
		}
	}

	// loop for all namespaces
	for _, namespace := range namespaces {
		pods, err := clientSet.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("There are %d pods in the %s namespace\n", len(pods.Items), namespace)

		for _, pod := range pods.Items {
			podDetail, _ := clientSet.CoreV1().Pods(namespace).Get(context.TODO(), pod.Name, metav1.GetOptions{})
			for _, container := range podDetail.Spec.Containers {
				containerImage := strings.Split(container.Image, ":")
				imageName = containerImage[0]
				if len(containerImage) == 2 {
					tag = containerImage[1]
				} else {
					tag = "noTag"
				}
				podData = append(podData, []string{podDetail.Name, container.Name, imageName, tag, string(podDetail.Status.Phase)})
			}
		}
		// set header of the table
		header := []string{"Pod name", "Container Name", "Image", "Tag", "Status"}
		printInfo(header, podData)
	}
	return nil
}
