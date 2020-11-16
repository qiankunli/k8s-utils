package main

import (
	"fmt"
	"github.com/k8s-utils/pod"
)

func main() {
	// only for test
	fmt.Println(pod.IsPodReady(nil))
}
