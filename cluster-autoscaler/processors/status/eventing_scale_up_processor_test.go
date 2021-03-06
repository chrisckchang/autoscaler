/*
Copyright 2018 The Kubernetes Authors.

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

package status

import (
	"strings"
	"testing"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/utils/nodegroupset"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	kube_record "k8s.io/client-go/tools/record"

	"github.com/stretchr/testify/assert"
)

func TestEventingScaleUpStatusProcessor(t *testing.T) {
	p := &EventingScaleUpStatusProcessor{}
	p1 := BuildTestPod("p1", 0, 0)
	p2 := BuildTestPod("p2", 0, 0)
	p3 := BuildTestPod("p3", 0, 0)

	testCases := []struct {
		caseName            string
		state               *ScaleUpStatus
		expectedTriggered   int
		expectedNoTriggered int
	}{
		{
			caseName: "No scale up",
			state: &ScaleUpStatus{
				ScaleUpInfos:            []nodegroupset.ScaleUpInfo{},
				PodsRemainUnschedulable: []*apiv1.Pod{p1, p2},
			},
			expectedNoTriggered: 2,
		},
		{
			caseName: "Scale up",
			state: &ScaleUpStatus{
				ScaleUpInfos:            []nodegroupset.ScaleUpInfo{{}},
				PodsTriggeredScaleUp:    []*apiv1.Pod{p3},
				PodsRemainUnschedulable: []*apiv1.Pod{p1, p2},
			},
			expectedTriggered:   1,
			expectedNoTriggered: 2,
		},
	}

	for _, tc := range testCases {
		fakeRecorder := kube_record.NewFakeRecorder(5)
		context := &context.AutoscalingContext{
			Recorder: fakeRecorder,
		}
		p.Process(context, tc.state)
		triggered := 0
		noTriggered := 0
		for eventsLeft := true; eventsLeft; {
			select {
			case event := <-fakeRecorder.Events:
				if strings.Contains(event, "TriggeredScaleUp") {
					triggered += 1
				} else if strings.Contains(event, "NotTriggerScaleUp") {
					noTriggered += 1
				} else {
					t.Fatalf("Test case '%v' failed. Unexpected event %v", tc.caseName, event)
				}
			default:
				eventsLeft = false
			}
		}
		assert.Equal(t, tc.expectedTriggered, triggered, "Test case '%v' failed.", tc.caseName)
		assert.Equal(t, tc.expectedNoTriggered, noTriggered, "Test case '%v' failed.", tc.caseName)
	}
}
