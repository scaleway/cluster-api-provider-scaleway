package v1alpha1

import (
	"errors"
	"reflect"
	unsafe "unsafe"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apimachineryconversion "k8s.io/apimachinery/pkg/conversion"
	"k8s.io/utils/ptr"
	clusterv1beta1 "sigs.k8s.io/cluster-api/api/core/v1beta1" //nolint:staticcheck
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/conversion"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha2"
)

// ConvertTo converts this ScalewayCluster (v1alpha1) to the Hub version (v1alpha2).
func (src *ScalewayCluster) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*infrav1.ScalewayCluster)

	return Convert_v1alpha1_ScalewayCluster_To_v1alpha2_ScalewayCluster(src, dst, nil)
}

// ConvertFrom converts the Hub version (v1alpha2) to this ScalewayCluster (v1alpha1).
func (dst *ScalewayCluster) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*infrav1.ScalewayCluster)

	return Convert_v1alpha2_ScalewayCluster_To_v1alpha1_ScalewayCluster(src, dst, nil)
}

// ConvertTo converts this ScalewayClusterTemplate (v1alpha1) to the Hub version (v1alpha2).
func (src *ScalewayClusterTemplate) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*infrav1.ScalewayClusterTemplate)

	return Convert_v1alpha1_ScalewayClusterTemplate_To_v1alpha2_ScalewayClusterTemplate(src, dst, nil)
}

// ConvertFrom converts the Hub version (v1alpha2) to this ScalewayClusterTemplate (v1alpha1).
func (dst *ScalewayClusterTemplate) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*infrav1.ScalewayClusterTemplate)

	return Convert_v1alpha2_ScalewayClusterTemplate_To_v1alpha1_ScalewayClusterTemplate(src, dst, nil)
}

// ConvertTo converts this ScalewayMachine (v1alpha1) to the Hub version (v1alpha2).
func (src *ScalewayMachine) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*infrav1.ScalewayMachine)

	return Convert_v1alpha1_ScalewayMachine_To_v1alpha2_ScalewayMachine(src, dst, nil)
}

// ConvertFrom converts the Hub version (v1alpha2) to this ScalewayMachine (v1alpha1).
func (dst *ScalewayMachine) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*infrav1.ScalewayMachine)

	return Convert_v1alpha2_ScalewayMachine_To_v1alpha1_ScalewayMachine(src, dst, nil)
}

// ConvertTo converts this ScalewayMachineTemplate (v1alpha1) to the Hub version (v1alpha2).
func (src *ScalewayMachineTemplate) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*infrav1.ScalewayMachineTemplate)

	return Convert_v1alpha1_ScalewayMachineTemplate_To_v1alpha2_ScalewayMachineTemplate(src, dst, nil)
}

// ConvertFrom converts the Hub version (v1alpha2) to this ScalewayMachineTemplate (v1alpha1).
func (dst *ScalewayMachineTemplate) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*infrav1.ScalewayMachineTemplate)

	return Convert_v1alpha2_ScalewayMachineTemplate_To_v1alpha1_ScalewayMachineTemplate(src, dst, nil)
}

// ConvertTo converts this ScalewayManagedCluster (v1alpha1) to the Hub version (v1alpha2).
func (src *ScalewayManagedCluster) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*infrav1.ScalewayManagedCluster)

	return Convert_v1alpha1_ScalewayManagedCluster_To_v1alpha2_ScalewayManagedCluster(src, dst, nil)
}

// ConvertFrom converts the Hub version (v1alpha2) to this ScalewayManagedCluster (v1alpha1).
func (dst *ScalewayManagedCluster) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*infrav1.ScalewayManagedCluster)

	return Convert_v1alpha2_ScalewayManagedCluster_To_v1alpha1_ScalewayManagedCluster(src, dst, nil)
}

// ConvertTo converts this ScalewayManagedControlPlane (v1alpha1) to the Hub version (v1alpha2).
func (src *ScalewayManagedControlPlane) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*infrav1.ScalewayManagedControlPlane)

	return Convert_v1alpha1_ScalewayManagedControlPlane_To_v1alpha2_ScalewayManagedControlPlane(src, dst, nil)
}

// ConvertFrom converts the Hub version (v1alpha2) to this ScalewayManagedControlPlane (v1alpha1).
func (dst *ScalewayManagedControlPlane) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*infrav1.ScalewayManagedControlPlane)

	return Convert_v1alpha2_ScalewayManagedControlPlane_To_v1alpha1_ScalewayManagedControlPlane(src, dst, nil)
}

// ConvertTo converts this ScalewayManagedMachinePool (v1alpha1) to the Hub version (v1alpha2).
func (src *ScalewayManagedMachinePool) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*infrav1.ScalewayManagedMachinePool)

	return Convert_v1alpha1_ScalewayManagedMachinePool_To_v1alpha2_ScalewayManagedMachinePool(src, dst, nil)
}

// ConvertFrom converts the Hub version (v1alpha2) to this ScalewayManagedMachinePool (v1alpha1).
func (dst *ScalewayManagedMachinePool) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*infrav1.ScalewayManagedMachinePool)

	return Convert_v1alpha2_ScalewayManagedMachinePool_To_v1alpha1_ScalewayManagedMachinePool(src, dst, nil)
}

func Convert_v1beta1_APIEndpoint_To_v1beta2_APIEndpoint(in *clusterv1beta1.APIEndpoint, out *clusterv1.APIEndpoint, s apimachineryconversion.Scope) error {
	return clusterv1beta1.Convert_v1beta1_APIEndpoint_To_v1beta2_APIEndpoint(in, out, s)
}

func Convert_v1beta2_APIEndpoint_To_v1beta1_APIEndpoint(in *clusterv1.APIEndpoint, out *clusterv1beta1.APIEndpoint, s apimachineryconversion.Scope) error {
	return clusterv1beta1.Convert_v1beta2_APIEndpoint_To_v1beta1_APIEndpoint(in, out, s)
}

func Convert_v1_ObjectMeta_To_v1beta2_ObjectMeta(in *metav1.ObjectMeta, out *clusterv1.ObjectMeta, s apimachineryconversion.Scope) error {
	out.Annotations = in.Annotations
	out.Labels = in.Labels

	return nil
}

func Convert_v1beta2_ObjectMeta_To_v1_ObjectMeta(in *clusterv1.ObjectMeta, out *metav1.ObjectMeta, s apimachineryconversion.Scope) error {
	out.Annotations = in.Annotations
	out.Labels = in.Labels

	return nil
}

func Convert_v1alpha1_PrivateNetworkSpec_To_v1alpha2_PrivateNetworkSpec(in *PrivateNetworkSpec, out *infrav1.PrivateNetworkSpec, s apimachineryconversion.Scope) error {
	if in == nil {
		return nil
	}

	if err := autoConvert_v1alpha1_PrivateNetworkSpec_To_v1alpha2_PrivateNetworkSpec(in, out, s); err != nil {
		return err
	}

	out.ID = infrav1.UUID(ptr.Deref(in.ID, ""))
	out.VPCID = infrav1.UUID(ptr.Deref(in.VPCID, ""))
	out.Subnet = infrav1.CIDR(ptr.Deref(in.Subnet, ""))

	return nil
}

func Convert_v1alpha2_PrivateNetworkSpec_To_v1alpha1_PrivateNetworkSpec(in *infrav1.PrivateNetworkSpec, out *PrivateNetworkSpec, s apimachineryconversion.Scope) error {
	if err := autoConvert_v1alpha2_PrivateNetworkSpec_To_v1alpha1_PrivateNetworkSpec(in, out, s); err != nil {
		return err
	}

	out.ID = ptrIfNotZero(string(in.ID))
	out.VPCID = ptrIfNotZero(string(in.VPCID))
	out.Subnet = ptrIfNotZero(string(in.Subnet))

	return nil
}

func Convert_v1alpha1_ScalewayClusterSpec_To_v1alpha2_ScalewayClusterSpec(in *ScalewayClusterSpec, out *infrav1.ScalewayClusterSpec, s apimachineryconversion.Scope) error {
	if err := autoConvert_v1alpha1_ScalewayClusterSpec_To_v1alpha2_ScalewayClusterSpec(in, out, nil); err != nil {
		return err
	}

	if in.Network == nil {
		return nil
	}

	var privateLB bool

	if in.Network.ControlPlaneLoadBalancer != nil {
		out.Network.ControlPlaneLoadBalancer.Zone = infrav1.ScalewayZone(ptr.Deref(in.Network.ControlPlaneLoadBalancer.Zone, ""))
		out.Network.ControlPlaneLoadBalancer.Type = ptr.Deref(in.Network.ControlPlaneLoadBalancer.Type, "")
		out.Network.ControlPlaneLoadBalancer.IP = infrav1.IPv4(ptr.Deref(in.Network.ControlPlaneLoadBalancer.IP, ""))
		out.Network.ControlPlaneLoadBalancer.PrivateIP = infrav1.IPv4(ptr.Deref(in.Network.ControlPlaneLoadBalancer.PrivateIP, ""))
		out.Network.ControlPlaneLoadBalancer.AllowedRanges = *(*[]infrav1.CIDR)(unsafe.Pointer(&in.Network.ControlPlaneLoadBalancer.AllowedRanges))
		out.Network.ControlPlaneLoadBalancer.Private = in.Network.ControlPlaneLoadBalancer.Private
		privateLB = ptr.Deref(in.Network.ControlPlaneLoadBalancer.Private, false)
	}

	for _, lb := range in.Network.ControlPlaneExtraLoadBalancers {
		out.Network.ControlPlaneExtraLoadBalancers = append(
			out.Network.ControlPlaneExtraLoadBalancers,
			infrav1.LoadBalancer{
				Zone:      infrav1.ScalewayZone(ptr.Deref(lb.Zone, "")),
				Type:      ptr.Deref(lb.Type, ""),
				IP:        infrav1.IPv4(ptr.Deref(lb.IP, "")),
				PrivateIP: infrav1.IPv4(ptr.Deref(lb.PrivateIP, "")),
			},
		)
	}

	if in.Network.ControlPlaneDNS != nil {
		if privateLB {
			return errors.New("cluster with controlPlaneDNS and a private LB cannot be converted")
		}

		out.Network.ControlPlaneDNS.Domain = in.Network.ControlPlaneDNS.Domain
		out.Network.ControlPlaneDNS.Name = in.Network.ControlPlaneDNS.Name
	}

	// NOTE: ControlPlanePrivateDNS is merged into ControlPlaneDNS during conversion.
	if in.Network.ControlPlanePrivateDNS != nil {
		out.Network.ControlPlaneDNS.Name = in.Network.ControlPlanePrivateDNS.Name
	}

	if err := Convert_v1alpha1_PrivateNetworkSpec_To_v1alpha2_PrivateNetworkSpec(in.Network.PrivateNetwork, &out.Network.PrivateNetwork, s); err != nil {
		return err
	}

	for _, pgw := range in.Network.PublicGateways {
		out.Network.PublicGateways = append(out.Network.PublicGateways, infrav1.PublicGateway{
			Type: ptr.Deref(pgw.Type, ""),
			IP:   infrav1.IPv4(ptr.Deref(pgw.IP, "")),
			Zone: infrav1.ScalewayZone(ptr.Deref(pgw.Zone, "")),
		})
	}

	return nil
}

func Convert_v1alpha2_ScalewayClusterSpec_To_v1alpha1_ScalewayClusterSpec(in *infrav1.ScalewayClusterSpec, out *ScalewayClusterSpec, s apimachineryconversion.Scope) error {
	if err := autoConvert_v1alpha2_ScalewayClusterSpec_To_v1alpha1_ScalewayClusterSpec(in, out, s); err != nil {
		return err
	}

	// Exit early if in.Network is not defined.
	if reflect.DeepEqual(in.Network, infrav1.ScalewayClusterNetwork{}) {
		return nil
	}

	out.Network = &NetworkSpec{}

	// Used to determine if we should put v1alpha2 ControlPlaneDNS into v1alpha1 ControlPlaneDNS or ControlPlanePrivateDNS.
	var privateLB bool

	if !reflect.DeepEqual(in.Network.ControlPlaneLoadBalancer, infrav1.ControlPlaneLoadBalancer{}) {
		out.Network.ControlPlaneLoadBalancer = &ControlPlaneLoadBalancerSpec{
			Private: in.Network.ControlPlaneLoadBalancer.Private,
			LoadBalancerSpec: LoadBalancerSpec{
				Zone:      ptrIfNotZero(string(in.Network.ControlPlaneLoadBalancer.Zone)),
				Type:      ptrIfNotZero(in.Network.ControlPlaneLoadBalancer.Type),
				IP:        ptrIfNotZero(string(in.Network.ControlPlaneLoadBalancer.IP)),
				PrivateIP: ptrIfNotZero(string(in.Network.ControlPlaneLoadBalancer.PrivateIP)),
			},
			AllowedRanges: *(*[]CIDR)(unsafe.Pointer(&in.Network.ControlPlaneLoadBalancer.AllowedRanges)),
		}

		privateLB = ptr.Deref(in.Network.ControlPlaneLoadBalancer.Private, false)
	}

	// Extra LBs.
	for _, lb := range in.Network.ControlPlaneExtraLoadBalancers {
		out.Network.ControlPlaneExtraLoadBalancers = append(
			out.Network.ControlPlaneExtraLoadBalancers,
			LoadBalancerSpec{
				IP:        ptrIfNotZero(string(lb.IP)),
				Type:      ptrIfNotZero(lb.Type),
				PrivateIP: ptrIfNotZero(string(lb.PrivateIP)),
				Zone:      ptrIfNotZero(string(lb.Zone)),
			},
		)
	}

	if !reflect.DeepEqual(in.Network.ControlPlaneDNS, infrav1.ControlPlaneDNS{}) {
		if privateLB {
			out.Network.ControlPlanePrivateDNS = &ControlPlanePrivateDNSSpec{
				Name: in.Network.ControlPlaneDNS.Name,
			}
		} else {
			out.Network.ControlPlaneDNS = &ControlPlaneDNSSpec{
				Domain: in.Network.ControlPlaneDNS.Domain,
				Name:   in.Network.ControlPlaneDNS.Name,
			}
		}
	}

	if !reflect.DeepEqual(in.Network.PrivateNetwork, infrav1.PrivateNetworkSpec{}) {
		out.Network.PrivateNetwork = &PrivateNetworkSpec{}
		if err := Convert_v1alpha2_PrivateNetworkSpec_To_v1alpha1_PrivateNetworkSpec(&in.Network.PrivateNetwork, out.Network.PrivateNetwork, s); err != nil {
			return err
		}
	}

	for _, pgw := range in.Network.PublicGateways {
		out.Network.PublicGateways = append(out.Network.PublicGateways, PublicGatewaySpec{
			IP:   ptrIfNotZero(string(pgw.IP)),
			Type: ptrIfNotZero(pgw.Type),
			Zone: ptrIfNotZero(string(pgw.Zone)),
		})
	}

	return nil
}

func Convert_v1alpha1_ScalewayClusterStatus_To_v1alpha2_ScalewayClusterStatus(in *ScalewayClusterStatus, out *infrav1.ScalewayClusterStatus, s apimachineryconversion.Scope) error {
	if err := autoConvert_v1alpha1_ScalewayClusterStatus_To_v1alpha2_ScalewayClusterStatus(in, out, s); err != nil {
		return err
	}

	for name, fd := range in.FailureDomains {
		out.FailureDomains = append(out.FailureDomains, clusterv1.FailureDomain{
			Name:         name,
			ControlPlane: &fd.ControlPlane,
			Attributes:   fd.Attributes,
		})
	}

	out.Initialization.Provisioned = &in.Ready

	if in.Network != nil {
		out.Network.VPCID = infrav1.UUID(ptr.Deref(in.Network.VPCID, ""))
		out.Network.PrivateNetworkID = infrav1.UUID(ptr.Deref(in.Network.PrivateNetworkID, ""))
		out.Network.PublicGatewayIDs = *(*[]infrav1.UUID)(unsafe.Pointer(&in.Network.PublicGatewayIDs))
		out.Network.LoadBalancerIP = infrav1.IPv4(ptr.Deref(in.Network.LoadBalancerIP, ""))
		out.Network.ExtraLoadBalancerIPs = *(*[]infrav1.IPv4)(unsafe.Pointer(&in.Network.ExtraLoadBalancerIPs))
	}

	return nil
}

func Convert_v1alpha2_ScalewayClusterStatus_To_v1alpha1_ScalewayClusterStatus(in *infrav1.ScalewayClusterStatus, out *ScalewayClusterStatus, s apimachineryconversion.Scope) error {
	if err := autoConvert_v1alpha2_ScalewayClusterStatus_To_v1alpha1_ScalewayClusterStatus(in, out, s); err != nil {
		return err
	}

	for _, fd := range in.FailureDomains {
		if out.FailureDomains == nil {
			out.FailureDomains = make(clusterv1beta1.FailureDomains)
		}

		out.FailureDomains[fd.Name] = clusterv1beta1.FailureDomainSpec{
			ControlPlane: ptr.Deref(fd.ControlPlane, false),
			Attributes:   fd.Attributes,
		}
	}

	out.Ready = ptr.Deref(in.Initialization.Provisioned, false)

	if !reflect.DeepEqual(in.Network, infrav1.ScalewayClusterNetworkStatus{}) {
		out.Network = &NetworkStatus{
			VPCID:                ptrIfNotZero(string(in.Network.VPCID)),
			PrivateNetworkID:     ptrIfNotZero(string(in.Network.PrivateNetworkID)),
			LoadBalancerIP:       ptrIfNotZero(string(in.Network.LoadBalancerIP)),
			PublicGatewayIDs:     *(*[]string)(unsafe.Pointer(&in.Network.PublicGatewayIDs)),
			ExtraLoadBalancerIPs: *(*[]string)(unsafe.Pointer(&in.Network.ExtraLoadBalancerIPs)),
		}
	}

	return nil
}

func Convert_v1alpha1_ScalewayMachineSpec_To_v1alpha2_ScalewayMachineSpec(in *ScalewayMachineSpec, out *infrav1.ScalewayMachineSpec, s apimachineryconversion.Scope) error {
	if err := autoConvert_v1alpha1_ScalewayMachineSpec_To_v1alpha2_ScalewayMachineSpec(in, out, s); err != nil {
		return err
	}

	if in.RootVolume != nil {
		out.RootVolume.IOPS = ptr.Deref(in.RootVolume.IOPS, 0)
		out.RootVolume.Size = ptr.Deref(in.RootVolume.Size, 0)
		out.RootVolume.Type = ptr.Deref(in.RootVolume.Type, "")
	}

	if in.PublicNetwork != nil {
		out.PublicNetwork.EnableIPv4 = in.PublicNetwork.EnableIPv4
		out.PublicNetwork.EnableIPv6 = in.PublicNetwork.EnableIPv6
	}

	if in.PlacementGroup != nil {
		out.PlacementGroup = infrav1.IDOrName{
			ID:   infrav1.UUID(ptr.Deref(in.PlacementGroup.ID, "")),
			Name: ptr.Deref(in.PlacementGroup.Name, ""),
		}
	}

	if in.SecurityGroup != nil {
		out.SecurityGroup = infrav1.IDOrName{
			ID:   infrav1.UUID(ptr.Deref(in.SecurityGroup.ID, "")),
			Name: ptr.Deref(in.SecurityGroup.Name, ""),
		}
	}

	return nil
}

func Convert_v1alpha2_ScalewayMachineSpec_To_v1alpha1_ScalewayMachineSpec(in *infrav1.ScalewayMachineSpec, out *ScalewayMachineSpec, s apimachineryconversion.Scope) error {
	if err := autoConvert_v1alpha2_ScalewayMachineSpec_To_v1alpha1_ScalewayMachineSpec(in, out, s); err != nil {
		return err
	}

	if !reflect.DeepEqual(in.RootVolume, infrav1.RootVolume{}) {
		out.RootVolume = &RootVolumeSpec{
			Size: ptrIfNotZero(in.RootVolume.Size),
			Type: ptrIfNotZero(in.RootVolume.Type),
			IOPS: ptrIfNotZero(in.RootVolume.IOPS),
		}
	}

	if !reflect.DeepEqual(in.PublicNetwork, infrav1.PublicNetwork{}) {
		out.PublicNetwork = &PublicNetworkSpec{
			EnableIPv4: in.PublicNetwork.EnableIPv4,
			EnableIPv6: in.PublicNetwork.EnableIPv6,
		}
	}

	if !reflect.DeepEqual(in.PlacementGroup, infrav1.IDOrName{}) {
		out.PlacementGroup = &PlacementGroupSpec{
			ID:   ptrIfNotZero(string(in.PlacementGroup.ID)),
			Name: ptrIfNotZero(in.PlacementGroup.Name),
		}
	}

	if !reflect.DeepEqual(in.SecurityGroup, infrav1.IDOrName{}) {
		out.SecurityGroup = &SecurityGroupSpec{
			ID:   ptrIfNotZero(string(in.SecurityGroup.ID)),
			Name: ptrIfNotZero(in.SecurityGroup.Name),
		}
	}

	return nil
}
func Convert_v1alpha1_ScalewayMachineStatus_To_v1alpha2_ScalewayMachineStatus(in *ScalewayMachineStatus, out *infrav1.ScalewayMachineStatus, s apimachineryconversion.Scope) error {
	if err := autoConvert_v1alpha1_ScalewayMachineStatus_To_v1alpha2_ScalewayMachineStatus(in, out, s); err != nil {
		return err
	}

	out.Initialization.Provisioned = &in.Ready

	return nil
}

func Convert_v1alpha2_ScalewayMachineStatus_To_v1alpha1_ScalewayMachineStatus(in *infrav1.ScalewayMachineStatus, out *ScalewayMachineStatus, s apimachineryconversion.Scope) error {
	if err := autoConvert_v1alpha2_ScalewayMachineStatus_To_v1alpha1_ScalewayMachineStatus(in, out, s); err != nil {
		return err
	}

	out.Ready = ptr.Deref(in.Initialization.Provisioned, false)

	return nil
}

func Convert_v1alpha1_ScalewayManagedClusterSpec_To_v1alpha2_ScalewayManagedClusterSpec(in *ScalewayManagedClusterSpec, out *infrav1.ScalewayManagedClusterSpec, s apimachineryconversion.Scope) error {
	if err := autoConvert_v1alpha1_ScalewayManagedClusterSpec_To_v1alpha2_ScalewayManagedClusterSpec(in, out, s); err != nil {
		return err
	}

	if in.Network == nil {
		return nil
	}

	if in.Network.PrivateNetwork != nil {
		out.Network.PrivateNetwork.ID = infrav1.UUID(ptr.Deref(in.Network.PrivateNetwork.ID, ""))
		out.Network.PrivateNetwork.Subnet = infrav1.CIDR(ptr.Deref(in.Network.PrivateNetwork.Subnet, ""))
		out.Network.PrivateNetwork.VPCID = infrav1.UUID(ptr.Deref(in.Network.PrivateNetwork.VPCID, ""))
	}

	for _, pgw := range in.Network.PublicGateways {
		out.Network.PublicGateways = append(out.Network.PublicGateways, infrav1.PublicGateway{
			Type: ptr.Deref(pgw.Type, ""),
			IP:   infrav1.IPv4(ptr.Deref(pgw.IP, "")),
			Zone: infrav1.ScalewayZone(ptr.Deref(pgw.Zone, "")),
		})
	}

	return nil
}

func Convert_v1alpha2_ScalewayManagedClusterSpec_To_v1alpha1_ScalewayManagedClusterSpec(in *infrav1.ScalewayManagedClusterSpec, out *ScalewayManagedClusterSpec, s apimachineryconversion.Scope) error {
	if err := autoConvert_v1alpha2_ScalewayManagedClusterSpec_To_v1alpha1_ScalewayManagedClusterSpec(in, out, s); err != nil {
		return err
	}

	if reflect.DeepEqual(in.Network, infrav1.ScalewayManagedClusterNetwork{}) {
		return nil
	}

	out.Network = &ManagedNetworkSpec{}

	if !reflect.DeepEqual(in.Network.PrivateNetwork, infrav1.PrivateNetwork{}) {
		out.Network.PrivateNetwork = &PrivateNetworkParams{
			ID:     ptrIfNotZero(string(in.Network.PrivateNetwork.ID)),
			VPCID:  ptrIfNotZero(string(in.Network.PrivateNetwork.VPCID)),
			Subnet: ptrIfNotZero(string(in.Network.PrivateNetwork.Subnet)),
		}
	}

	for _, pgw := range in.Network.PublicGateways {
		out.Network.PublicGateways = append(out.Network.PublicGateways, PublicGatewaySpec{
			Type: ptrIfNotZero(pgw.Type),
			IP:   ptrIfNotZero(string(pgw.IP)),
			Zone: ptrIfNotZero(string(pgw.Zone)),
		})
	}

	return nil
}

func Convert_v1alpha1_ScalewayManagedClusterStatus_To_v1alpha2_ScalewayManagedClusterStatus(in *ScalewayManagedClusterStatus, out *infrav1.ScalewayManagedClusterStatus, s apimachineryconversion.Scope) error {
	if err := autoConvert_v1alpha1_ScalewayManagedClusterStatus_To_v1alpha2_ScalewayManagedClusterStatus(in, out, s); err != nil {
		return err
	}

	out.Initialization.Provisioned = &in.Ready

	if in.Network != nil {
		out.Network.PrivateNetworkID = infrav1.UUID(ptr.Deref(in.Network.PrivateNetworkID, ""))
	}

	return nil
}

func Convert_v1alpha2_ScalewayManagedClusterStatus_To_v1alpha1_ScalewayManagedClusterStatus(in *infrav1.ScalewayManagedClusterStatus, out *ScalewayManagedClusterStatus, s apimachineryconversion.Scope) error {
	if err := autoConvert_v1alpha2_ScalewayManagedClusterStatus_To_v1alpha1_ScalewayManagedClusterStatus(in, out, s); err != nil {
		return err
	}

	out.Ready = ptr.Deref(in.Initialization.Provisioned, false)

	if in.Network.PrivateNetworkID != "" {
		out.Network = &ManagedNetworkStatus{
			PrivateNetworkID: ptrIfNotZero(string(in.Network.PrivateNetworkID)),
		}
	}

	return nil
}

func Convert_v1alpha1_ScalewayManagedControlPlaneSpec_To_v1alpha2_ScalewayManagedControlPlaneSpec(in *ScalewayManagedControlPlaneSpec, out *infrav1.ScalewayManagedControlPlaneSpec, s apimachineryconversion.Scope) error {
	if err := autoConvert_v1alpha1_ScalewayManagedControlPlaneSpec_To_v1alpha2_ScalewayManagedControlPlaneSpec(in, out, s); err != nil {
		return err
	}

	if in.Autoscaler != nil {
		out.Autoscaler.ScaleDownDisabled = in.Autoscaler.ScaleDownDisabled
		out.Autoscaler.ScaleDownDelayAfterAdd = ptr.Deref(in.Autoscaler.ScaleDownDelayAfterAdd, "")
		out.Autoscaler.Estimator = ptr.Deref(in.Autoscaler.Estimator, "")
		out.Autoscaler.Expander = ptr.Deref(in.Autoscaler.Expander, "")
		out.Autoscaler.IgnoreDaemonsetsUtilization = in.Autoscaler.IgnoreDaemonsetsUtilization
		out.Autoscaler.BalanceSimilarNodeGroups = in.Autoscaler.BalanceSimilarNodeGroups
		out.Autoscaler.ExpendablePodsPriorityCutoff = in.Autoscaler.ExpendablePodsPriorityCutoff
		out.Autoscaler.ScaleDownUnneededTime = ptr.Deref(in.Autoscaler.ScaleDownUnneededTime, "")
		out.Autoscaler.ScaleDownUtilizationThreshold = ptr.Deref(in.Autoscaler.ScaleDownUtilizationThreshold, "")
		out.Autoscaler.MaxGracefulTerminationSec = ptr.Deref(in.Autoscaler.MaxGracefulTerminationSec, 0)
	}

	if in.AutoUpgrade != nil {
		out.AutoUpgrade.Enabled = &in.AutoUpgrade.Enabled

		if in.AutoUpgrade.MaintenanceWindow != nil {
			out.AutoUpgrade.MaintenanceWindow.Day = ptr.Deref(in.AutoUpgrade.MaintenanceWindow.Day, "")
			out.AutoUpgrade.MaintenanceWindow.StartHour = in.AutoUpgrade.MaintenanceWindow.StartHour
		}
	}

	if in.OpenIDConnect != nil {
		out.OpenIDConnect.IssuerURL = in.OpenIDConnect.IssuerURL
		out.OpenIDConnect.ClientID = in.OpenIDConnect.ClientID
		out.OpenIDConnect.UsernameClaim = ptr.Deref(in.OpenIDConnect.UsernameClaim, "")
		out.OpenIDConnect.UsernamePrefix = ptr.Deref(in.OpenIDConnect.UsernamePrefix, "")
		out.OpenIDConnect.GroupsClaim = in.OpenIDConnect.GroupsClaim
		out.OpenIDConnect.GroupsPrefix = ptr.Deref(in.OpenIDConnect.GroupsPrefix, "")
		out.OpenIDConnect.RequiredClaim = in.OpenIDConnect.RequiredClaim
	}

	if in.OnDelete != nil {
		out.OnDelete.WithAdditionalResources = in.OnDelete.WithAdditionalResources
	}

	return nil
}

func Convert_v1alpha2_ScalewayManagedControlPlaneSpec_To_v1alpha1_ScalewayManagedControlPlaneSpec(in *infrav1.ScalewayManagedControlPlaneSpec, out *ScalewayManagedControlPlaneSpec, s apimachineryconversion.Scope) error {
	if err := autoConvert_v1alpha2_ScalewayManagedControlPlaneSpec_To_v1alpha1_ScalewayManagedControlPlaneSpec(in, out, s); err != nil {
		return err
	}

	if !reflect.DeepEqual(in.Autoscaler, infrav1.Autoscaler{}) {
		out.Autoscaler = &AutoscalerSpec{
			ScaleDownDisabled:             in.Autoscaler.ScaleDownDisabled,
			ScaleDownDelayAfterAdd:        ptrIfNotZero(in.Autoscaler.ScaleDownDelayAfterAdd),
			Estimator:                     ptrIfNotZero(in.Autoscaler.Estimator),
			Expander:                      ptrIfNotZero(in.Autoscaler.Expander),
			IgnoreDaemonsetsUtilization:   in.Autoscaler.IgnoreDaemonsetsUtilization,
			BalanceSimilarNodeGroups:      in.Autoscaler.BalanceSimilarNodeGroups,
			ExpendablePodsPriorityCutoff:  in.Autoscaler.ExpendablePodsPriorityCutoff,
			ScaleDownUnneededTime:         ptrIfNotZero(in.Autoscaler.ScaleDownUnneededTime),
			ScaleDownUtilizationThreshold: ptrIfNotZero(in.Autoscaler.ScaleDownUtilizationThreshold),
			MaxGracefulTerminationSec:     ptrIfNotZero(in.Autoscaler.MaxGracefulTerminationSec),
		}
	}

	if !reflect.DeepEqual(in.AutoUpgrade, infrav1.AutoUpgrade{}) {
		out.AutoUpgrade = &AutoUpgradeSpec{
			Enabled: ptr.Deref(in.AutoUpgrade.Enabled, false),
		}

		if !reflect.DeepEqual(in.AutoUpgrade.MaintenanceWindow, infrav1.MaintenanceWindow{}) {
			out.AutoUpgrade.MaintenanceWindow = &MaintenanceWindowSpec{
				StartHour: in.AutoUpgrade.MaintenanceWindow.StartHour,
				Day:       ptrIfNotZero(in.AutoUpgrade.MaintenanceWindow.Day),
			}
		}
	}

	if !reflect.DeepEqual(in.OpenIDConnect, infrav1.OpenIDConnect{}) {
		out.OpenIDConnect = &OpenIDConnectSpec{
			IssuerURL:      in.OpenIDConnect.IssuerURL,
			ClientID:       in.OpenIDConnect.ClientID,
			UsernameClaim:  ptrIfNotZero(in.OpenIDConnect.UsernameClaim),
			UsernamePrefix: ptrIfNotZero(in.OpenIDConnect.UsernamePrefix),
			GroupsClaim:    in.OpenIDConnect.GroupsClaim,
			GroupsPrefix:   ptrIfNotZero(in.OpenIDConnect.GroupsPrefix),
			RequiredClaim:  in.OpenIDConnect.RequiredClaim,
		}
	}

	if !reflect.DeepEqual(in.OnDelete, infrav1.OnDelete{}) {
		out.OnDelete = &OnDeleteSpec{
			WithAdditionalResources: in.OnDelete.WithAdditionalResources,
		}
	}

	return nil
}

func Convert_v1alpha1_ScalewayManagedControlPlaneStatus_To_v1alpha2_ScalewayManagedControlPlaneStatus(in *ScalewayManagedControlPlaneStatus, out *infrav1.ScalewayManagedControlPlaneStatus, s apimachineryconversion.Scope) error {
	if err := autoConvert_v1alpha1_ScalewayManagedControlPlaneStatus_To_v1alpha2_ScalewayManagedControlPlaneStatus(in, out, s); err != nil {
		return err
	}

	// Drop in.Ready

	out.Initialization.ControlPlaneInitialized = &in.Initialized

	return nil
}

func Convert_v1alpha2_ScalewayManagedControlPlaneStatus_To_v1alpha1_ScalewayManagedControlPlaneStatus(in *infrav1.ScalewayManagedControlPlaneStatus, out *ScalewayManagedControlPlaneStatus, s apimachineryconversion.Scope) error {
	if err := autoConvert_v1alpha2_ScalewayManagedControlPlaneStatus_To_v1alpha1_ScalewayManagedControlPlaneStatus(in, out, s); err != nil {
		return err
	}

	out.Ready = ptr.Deref(in.Initialization.ControlPlaneInitialized, false)
	out.Initialized = ptr.Deref(in.Initialization.ControlPlaneInitialized, false)

	return nil
}

func Convert_v1alpha1_ScalewayManagedMachinePoolSpec_To_v1alpha2_ScalewayManagedMachinePoolSpec(in *ScalewayManagedMachinePoolSpec, out *infrav1.ScalewayManagedMachinePoolSpec, s apimachineryconversion.Scope) error {
	if err := autoConvert_v1alpha1_ScalewayManagedMachinePoolSpec_To_v1alpha2_ScalewayManagedMachinePoolSpec(in, out, s); err != nil {
		return err
	}

	out.PlacementGroupID = infrav1.UUID(ptr.Deref(in.PlacementGroupID, ""))
	out.SecurityGroupID = infrav1.UUID(ptr.Deref(in.SecurityGroupID, ""))

	if in.Scaling != nil {
		out.Scaling.Autoscaling = in.Scaling.Autoscaling
		out.Scaling.MinSize = in.Scaling.MinSize
		out.Scaling.MaxSize = in.Scaling.MaxSize
	}

	if in.UpgradePolicy != nil {
		out.UpgradePolicy.MaxSurge = in.UpgradePolicy.MaxSurge
		out.UpgradePolicy.MaxUnavailable = in.UpgradePolicy.MaxUnavailable
	}

	return nil
}

func Convert_v1alpha2_ScalewayManagedMachinePoolSpec_To_v1alpha1_ScalewayManagedMachinePoolSpec(in *infrav1.ScalewayManagedMachinePoolSpec, out *ScalewayManagedMachinePoolSpec, s apimachineryconversion.Scope) error {
	if err := autoConvert_v1alpha2_ScalewayManagedMachinePoolSpec_To_v1alpha1_ScalewayManagedMachinePoolSpec(in, out, s); err != nil {
		return err
	}

	out.PlacementGroupID = ptrIfNotZero(string(in.PlacementGroupID))
	out.SecurityGroupID = ptrIfNotZero(string(in.SecurityGroupID))

	if !reflect.DeepEqual(in.Scaling, infrav1.Scaling{}) {
		out.Scaling = &ScalingSpec{
			Autoscaling: in.Scaling.Autoscaling,
			MinSize:     in.Scaling.MinSize,
			MaxSize:     in.Scaling.MaxSize,
		}
	}

	if !reflect.DeepEqual(in.UpgradePolicy, infrav1.UpgradePolicy{}) {
		out.UpgradePolicy = &UpgradePolicySpec{
			MaxSurge:       in.UpgradePolicy.MaxSurge,
			MaxUnavailable: in.UpgradePolicy.MaxUnavailable,
		}
	}

	return nil
}

func Convert_v1alpha1_ScalewayManagedMachinePoolStatus_To_v1alpha2_ScalewayManagedMachinePoolStatus(in *ScalewayManagedMachinePoolStatus, out *infrav1.ScalewayManagedMachinePoolStatus, s apimachineryconversion.Scope) error {
	if err := autoConvert_v1alpha1_ScalewayManagedMachinePoolStatus_To_v1alpha2_ScalewayManagedMachinePoolStatus(in, out, s); err != nil {
		return err
	}

	out.Initialization.Provisioned = &in.Ready

	return nil
}
func Convert_v1alpha2_ScalewayManagedMachinePoolStatus_To_v1alpha1_ScalewayManagedMachinePoolStatus(in *infrav1.ScalewayManagedMachinePoolStatus, out *ScalewayManagedMachinePoolStatus, s apimachineryconversion.Scope) error {
	if err := autoConvert_v1alpha2_ScalewayManagedMachinePoolStatus_To_v1alpha1_ScalewayManagedMachinePoolStatus(in, out, s); err != nil {
		return err
	}

	out.Ready = ptr.Deref(in.Initialization.Provisioned, false)

	return nil
}

func Convert_v1alpha1_ImageSpec_To_v1alpha2_Image(in *ImageSpec, out *infrav1.Image, s apimachineryconversion.Scope) error {
	out.ID = infrav1.UUID(ptr.Deref(in.ID, ""))
	out.Name = ptr.Deref(in.Name, "")
	out.Label = ptr.Deref(in.Label, "")

	return nil
}
func Convert_v1alpha2_Image_To_v1alpha1_ImageSpec(in *infrav1.Image, out *ImageSpec, s apimachineryconversion.Scope) error {
	out.ID = ptrIfNotZero(string(in.ID))
	out.Label = ptrIfNotZero(in.Label)
	out.Name = ptrIfNotZero(in.Name)

	return nil
}

// ptrIfNotZero returns a pointer to v if v is NOT a zero value. Otherwise, it returns nil.
func ptrIfNotZero[T comparable](v T) *T {
	var e T

	if v == e {
		return nil
	}

	return &v
}
