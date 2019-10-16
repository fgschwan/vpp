// Copyright (c) 2018 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	telemetryv1 "github.com/contiv/vpp/plugins/crd/pkg/apis/telemetry/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeTelemetryReports implements TelemetryReportInterface
type FakeTelemetryReports struct {
	Fake *FakeTelemetryV1
	ns   string
}

var telemetryreportsResource = schema.GroupVersionResource{Group: "telemetry.contiv.vpp", Version: "v1", Resource: "telemetryreports"}

var telemetryreportsKind = schema.GroupVersionKind{Group: "telemetry.contiv.vpp", Version: "v1", Kind: "TelemetryReport"}

// Get takes name of the telemetryReport, and returns the corresponding telemetryReport object, and an error if there is any.
func (c *FakeTelemetryReports) Get(name string, options v1.GetOptions) (result *telemetryv1.TelemetryReport, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(telemetryreportsResource, c.ns, name), &telemetryv1.TelemetryReport{})

	if obj == nil {
		return nil, err
	}
	return obj.(*telemetryv1.TelemetryReport), err
}

// List takes label and field selectors, and returns the list of TelemetryReports that match those selectors.
func (c *FakeTelemetryReports) List(opts v1.ListOptions) (result *telemetryv1.TelemetryReportList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(telemetryreportsResource, telemetryreportsKind, c.ns, opts), &telemetryv1.TelemetryReportList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &telemetryv1.TelemetryReportList{ListMeta: obj.(*telemetryv1.TelemetryReportList).ListMeta}
	for _, item := range obj.(*telemetryv1.TelemetryReportList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested telemetryReports.
func (c *FakeTelemetryReports) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(telemetryreportsResource, c.ns, opts))

}

// Create takes the representation of a telemetryReport and creates it.  Returns the server's representation of the telemetryReport, and an error, if there is any.
func (c *FakeTelemetryReports) Create(telemetryReport *telemetryv1.TelemetryReport) (result *telemetryv1.TelemetryReport, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(telemetryreportsResource, c.ns, telemetryReport), &telemetryv1.TelemetryReport{})

	if obj == nil {
		return nil, err
	}
	return obj.(*telemetryv1.TelemetryReport), err
}

// Update takes the representation of a telemetryReport and updates it. Returns the server's representation of the telemetryReport, and an error, if there is any.
func (c *FakeTelemetryReports) Update(telemetryReport *telemetryv1.TelemetryReport) (result *telemetryv1.TelemetryReport, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(telemetryreportsResource, c.ns, telemetryReport), &telemetryv1.TelemetryReport{})

	if obj == nil {
		return nil, err
	}
	return obj.(*telemetryv1.TelemetryReport), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeTelemetryReports) UpdateStatus(telemetryReport *telemetryv1.TelemetryReport) (*telemetryv1.TelemetryReport, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(telemetryreportsResource, "status", c.ns, telemetryReport), &telemetryv1.TelemetryReport{})

	if obj == nil {
		return nil, err
	}
	return obj.(*telemetryv1.TelemetryReport), err
}

// Delete takes name of the telemetryReport and deletes it. Returns an error if one occurs.
func (c *FakeTelemetryReports) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(telemetryreportsResource, c.ns, name), &telemetryv1.TelemetryReport{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeTelemetryReports) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(telemetryreportsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &telemetryv1.TelemetryReportList{})
	return err
}

// Patch applies the patch and returns the patched telemetryReport.
func (c *FakeTelemetryReports) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *telemetryv1.TelemetryReport, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(telemetryreportsResource, c.ns, name, pt, data, subresources...), &telemetryv1.TelemetryReport{})

	if obj == nil {
		return nil, err
	}
	return obj.(*telemetryv1.TelemetryReport), err
}
