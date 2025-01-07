package main

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime/pprof"
	"time"

	"github.com/rs/zerolog/log"
	kubernetes "github.com/trivago/go-kubernetes/v3"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func connectToKubernetes(kubeConfigPath string) (map[string]*kubernetes.Client, error) {
	clusters := make(map[string]*kubernetes.Client)

	log.Debug().Msgf("getting config from %s", kubeConfigPath)

	contexts, err := kubernetes.GetContextsFromConfig(kubeConfigPath)
	if err != nil {
		log.Error().Msg("failed to read contexts from kubeconfig")
		return clusters, err
	}

	numContexts := 0

	for _, context := range contexts {
		numContexts++
		if numContexts > 3 {
			break
		}
		remoteClient, err := kubernetes.NewClientUsingContext(kubeConfigPath, context)
		if err != nil {
			log.Error().Msgf("failed to create remote kubernetes client for context %s", context)
			continue
		}
		log.Debug().Msgf("created remote kubernetes client for context %s", context)
		clusters[context] = remoteClient
	}

	return clusters, nil
}

func main() {
	// Start the CPU profiler
	if os.Getenv("PROFILE_OUT") == "file" {
		f, err := os.Create(time.Now().Format("cpu_20060102_1504.pprof"))
		if err != nil {
			log.Error().Msg("could not create CPU profile")
			return
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Error().Msg("could not start CPU profile")
			return
		}
		defer pprof.StopCPUProfile()
	} else {

		// Start profiler as HTTP endpoint
		go func() {
			err := http.ListenAndServe("localhost:6060", nil)
			if err != nil {
				log.Error().Msg("failed to start pprof server")
			}
		}()
	}

	// Set up contexts
	kubeConfigPath := os.Getenv("KUBECONFIG")
	if kubeConfigPath == "" {
		kubeConfigPath = os.ExpandEnv("$HOME/.kube/config")
	}

	clusters, err := connectToKubernetes(kubeConfigPath)
	if err != nil {
		log.Fatal().Msg("failed to connect to any context")
		return
	}

	// List namespaces in each context
	namespaceGVR := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "namespaces",
	}

	for context, client := range clusters {
		namespaces, err := client.ListAllObjects(namespaceGVR, "", "")
		if err != nil {
			log.Error().Msgf("failed to list namespaces in context %s", context)
			continue
		}
		log.Info().Msgf("namespaces in context %s: %v", context, namespaces)
	}
}
