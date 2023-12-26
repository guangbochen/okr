//go:build !ignore_autogenerated

/*
Copyright 2023.

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

// Code generated by controller-gen. DO NOT EDIT.

package v1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KThreesAgentConfig) DeepCopyInto(out *KThreesAgentConfig) {
	*out = *in
	if in.NodeLabels != nil {
		in, out := &in.NodeLabels, &out.NodeLabels
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.NodeTaints != nil {
		in, out := &in.NodeTaints, &out.NodeTaints
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.KubeletArgs != nil {
		in, out := &in.KubeletArgs, &out.KubeletArgs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.KubeProxyArgs != nil {
		in, out := &in.KubeProxyArgs, &out.KubeProxyArgs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KThreesAgentConfig.
func (in *KThreesAgentConfig) DeepCopy() *KThreesAgentConfig {
	if in == nil {
		return nil
	}
	out := new(KThreesAgentConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KThreesConfigStatus) DeepCopyInto(out *KThreesConfigStatus) {
	*out = *in
	if in.BootstrapData != nil {
		in, out := &in.BootstrapData, &out.BootstrapData
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	if in.DataSecretName != nil {
		in, out := &in.DataSecretName, &out.DataSecretName
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KThreesConfigStatus.
func (in *KThreesConfigStatus) DeepCopy() *KThreesConfigStatus {
	if in == nil {
		return nil
	}
	out := new(KThreesConfigStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KThreesServerConfig) DeepCopyInto(out *KThreesServerConfig) {
	*out = *in
	if in.KubeAPIServerArgs != nil {
		in, out := &in.KubeAPIServerArgs, &out.KubeAPIServerArgs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.KubeControllerManagerArgs != nil {
		in, out := &in.KubeControllerManagerArgs, &out.KubeControllerManagerArgs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.KubeSchedulerArgs != nil {
		in, out := &in.KubeSchedulerArgs, &out.KubeSchedulerArgs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.TLSSan != nil {
		in, out := &in.TLSSan, &out.TLSSan
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.DisableComponents != nil {
		in, out := &in.DisableComponents, &out.DisableComponents
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KThreesServerConfig.
func (in *KThreesServerConfig) DeepCopy() *KThreesServerConfig {
	if in == nil {
		return nil
	}
	out := new(KThreesServerConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Ok3sConfig) DeepCopyInto(out *Ok3sConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Ok3sConfig.
func (in *Ok3sConfig) DeepCopy() *Ok3sConfig {
	if in == nil {
		return nil
	}
	out := new(Ok3sConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Ok3sConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Ok3sConfigList) DeepCopyInto(out *Ok3sConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Ok3sConfig, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Ok3sConfigList.
func (in *Ok3sConfigList) DeepCopy() *Ok3sConfigList {
	if in == nil {
		return nil
	}
	out := new(Ok3sConfigList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Ok3sConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Ok3sConfigSpec) DeepCopyInto(out *Ok3sConfigSpec) {
	*out = *in
	in.ServerConfig.DeepCopyInto(&out.ServerConfig)
	in.AgentConfig.DeepCopyInto(&out.AgentConfig)
	if in.PreK3sCommands != nil {
		in, out := &in.PreK3sCommands, &out.PreK3sCommands
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.PostK3sCommands != nil {
		in, out := &in.PostK3sCommands, &out.PostK3sCommands
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Ok3sConfigSpec.
func (in *Ok3sConfigSpec) DeepCopy() *Ok3sConfigSpec {
	if in == nil {
		return nil
	}
	out := new(Ok3sConfigSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Ok3sConfigStatus) DeepCopyInto(out *Ok3sConfigStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Ok3sConfigStatus.
func (in *Ok3sConfigStatus) DeepCopy() *Ok3sConfigStatus {
	if in == nil {
		return nil
	}
	out := new(Ok3sConfigStatus)
	in.DeepCopyInto(out)
	return out
}