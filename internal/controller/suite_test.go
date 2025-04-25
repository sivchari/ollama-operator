/*
Copyright 2025.

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
	"context"
	"os"
	"testing"
	"time"

	. "github.com/onsi/gomega"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/sivchari/ollama-operator/internal/test/envtest"
)

var (
	ctx = ctrl.SetupSignalHandler()
	env *envtest.Environment
)

func TestMain(m *testing.M) {
	setupReconcilers := func(_ context.Context, mgr ctrl.Manager) {
		err := (&ModelReconciler{
			Client:               mgr.GetClient(),
			Scheme:               mgr.GetScheme(),
			OllamaContainerImage: "ollama/ollama:latest",
		}).SetupWithManager(mgr)
		if err != nil {
			panic(err)
		}
	}

	SetDefaultEventuallyPollingInterval(100 * time.Millisecond)
	SetDefaultEventuallyTimeout(30 * time.Second)

	os.Exit(envtest.Run(ctx, envtest.Input{
		M:                m,
		SetupReconcilers: setupReconcilers,
		SetupEnv:         func(e *envtest.Environment) { env = e },
	}))
}
