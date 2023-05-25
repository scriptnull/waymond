package awsec2asg

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/knadh/koanf/v2"
	"github.com/scriptnull/waymond/internal/event"
	"github.com/scriptnull/waymond/internal/log"
	"github.com/scriptnull/waymond/internal/scaler"
)

const Type scaler.Type = "aws_ec2_asg"

type Scaler struct {
	ID           string `koanf:"id"`
	namespacedID string
	log          log.Logger

	AllowCreate          bool                  `koanf:"allow_create"`
	DisableScaleIn       *bool                 `koanf:"disable_scale_in"`
	DisableScaleOut      *bool                 `koanf:"disable_scale_out"`
	MinSize              *int64                `koanf:"min_size"`
	MaxSize              *int64                `koanf:"max_size"`
	CapacityRebalance    *bool                 `koanf:"capacity_rebalance"`
	DefaultCooldown      *int64                `koanf:"default_cooldown"`
	VpcZoneIdentifier    []string              `koanf:"vpc_zone_identifier"`
	PlacementGroup       *string               `koanf:"placement_group"`
	Tags                 []ASGTag              `koanf:"tags"`
	MixedInstancesPolicy *MixedInstancesPolicy `koanf:"mixed_instances_policy"`
}

type ASGTag struct {
	Key               *string `koanf:"key"`
	Value             *string `koanf:"value"`
	PropagateAtLaunch *bool   `koanf:"propagate_at_launch"`
}

type MixedInstancesPolicy struct {
	InstanceDistribution *InstanceDistribution `koanf:"instances_distribution"`

	LaunchTemplate *LaunchTemplate `koanf:"launch_template"`
}

type InstanceDistribution struct {
	OnDemandAllocationStrategy       *string `koanf:"on_demand_allocation_strategy"`
	OnDemandBaseCapacity             *int64  `koanf:"on_demand_base_capacity"`
	OnDemandPercentAboveBaseCapacity *int64  `koanf:"on_demand_percentage_above_base_capacity"`
	SpotAllocationStrategy           *string `koanf:"spot_allocation_strategy"`
	SpotInstancePools                *int64  `koanf:"spot_instance_pools"`
	SpotMaxPrice                     *string `koanf:"spot_max_price"`
}

type LaunchTemplate struct {
	LaunchTemplateSpecification *LaunchTemplateSpecification `koanf:"launch_template_specification"`
	Overrides                   []*LaunchTemplateOverrides   `koanf:"overrides"`
}

type LaunchTemplateSpecification struct {
	LaunchTemplateId   *string `koanf:"launch_template_id" json:"launch_template_id"`
	LaunchTemplateName *string `koanf:"launch_template_name" json:"launch_template_name"`
	Version            *string `koanf:"version" json:"version"`
}

type LaunchTemplateOverrides struct {
	// TODO: too much of struct nesting in InstanceRequirements
	// so add it here when we actually need it.
	// InstanceRequirements *InstanceRequirements `koanf:"instance_requirements"`

	InstanceType                *string                      `koanf:"instance_type" json:"instance_type"`
	LaunchTemplateSpecification *LaunchTemplateSpecification `koanf:"launch_template_specification" json:"launch_template_specification"`
	WeightedCapacity            *string                      `koanf:"weighted_capacity" json:"weighted_capacity"`
}

func (s *Scaler) Type() scaler.Type {
	return Type
}

func (s *Scaler) Register(ctx context.Context) error {
	sess, err := session.NewSession()
	if err != nil {
		return err
	}

	svc := autoscaling.New(sess)
	ec2svc := ec2.New(sess)

	event.B.Subscribe(fmt.Sprintf("%s.input", s.namespacedID), func(data []byte) {
		s.log.Verbose("start")

		s.log.Debugf("data: %+v\n", string(data))

		var inputData struct {
			ASGName                   string                       `json:"asg_name"`
			DesiredCount              int64                        `json:"desired_count"`
			MinSize                   *int64                       `json:"min_size"`
			MaxSize                   *int64                       `json:"max_size"`
			Tags                      []ASGTag                     `json:"tags"`
			BaseLaunchTemplate        *LaunchTemplateSpecification `json:"base_launch_template"`
			LaunchTemplateVersionOpts struct {
				AmiID *string `json:"ami_id"`
			} `json:"launch_template_version_options"`
			Overrides []LaunchTemplateOverrides `json:"launch_template_overrides"`
		}

		err := json.Unmarshal(data, &inputData)
		if err != nil {
			s.log.Errorf("error: %s \n", err)
			return
		}

		var maxRecords int64 = 1
		asgOutput, err := svc.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: []*string{&inputData.ASGName},
			MaxRecords:            &maxRecords,
		})
		if err != nil {
			s.log.Error("error while trying to find autoscaling group", err)
			return
		}

		s.log.Debug("asg", asgOutput)

		if len(asgOutput.AutoScalingGroups) == 0 {
			// autoscaling group is absent in EC2
			// so create one if allow_create is set to true
			// else error out and return

			if !s.AllowCreate {
				s.log.Error("ASG not found. Please set `allow_create` to true if you would like to create it via waymond")
				return
			}

			// ensure launch template and its versions are available for the autoscaling group to make use of
			var preferredLaunchTemplateVersion *ec2.LaunchTemplateVersion
			if inputData.LaunchTemplateVersionOpts.AmiID != nil {
				if inputData.BaseLaunchTemplate == nil {
					s.log.Error("'base_launch_template' is required input for the aws_ec2_asg scaler if 'launch_template_version_options' is mentioned \n")
					return
				}
				s.log.Verbosef("querying for a launch template version with AMI ID: %s \n", *inputData.LaunchTemplateVersionOpts.AmiID)
				imageId := "image-id"
				ltvs, err := ec2svc.DescribeLaunchTemplateVersions(&ec2.DescribeLaunchTemplateVersionsInput{
					LaunchTemplateId:   inputData.BaseLaunchTemplate.LaunchTemplateId,
					LaunchTemplateName: inputData.BaseLaunchTemplate.LaunchTemplateName,
					Filters: []*ec2.Filter{
						{
							Name:   &imageId,
							Values: []*string{inputData.LaunchTemplateVersionOpts.AmiID},
						},
					},
				})
				if err != nil {
					s.log.Errorf("error querying for a launch template version containing AMI ID (%s): %s \n", *inputData.LaunchTemplateVersionOpts.AmiID, err)
					return
				}
				if ltvs != nil && len(ltvs.LaunchTemplateVersions) > 0 {
					ltv := ltvs.LaunchTemplateVersions[0]
					s.log.Verbosef("found a launch template version containing the given AMI ID: %s, launch template version: %d \n", *inputData.LaunchTemplateVersionOpts.AmiID, *ltv.VersionNumber)
					preferredLaunchTemplateVersion = ltv
				} else {
					s.log.Verbosef("unable to find a launch template version containing the given AMI ID: %s, so creating one. \n", *inputData.LaunchTemplateVersionOpts.AmiID)
					ltv, err := ec2svc.CreateLaunchTemplateVersion(&ec2.CreateLaunchTemplateVersionInput{
						LaunchTemplateName: inputData.BaseLaunchTemplate.LaunchTemplateName,
						LaunchTemplateId:   inputData.BaseLaunchTemplate.LaunchTemplateId,
						SourceVersion:      inputData.BaseLaunchTemplate.Version,
						LaunchTemplateData: &ec2.RequestLaunchTemplateData{
							ImageId: inputData.LaunchTemplateVersionOpts.AmiID,
						},
					})
					if err != nil {
						s.log.Error("error while creating a launch template version: %s", err)
						return
					}
					s.log.Verbosef("created a launch template version containing the given AMI ID: %s, launch template version: %d \n", *inputData.LaunchTemplateVersionOpts.AmiID, ltv.LaunchTemplateVersion.VersionNumber)
					preferredLaunchTemplateVersion = ltv.LaunchTemplateVersion
				}
			}

			s.log.Verbose("creating a new ASG")
			createAsgInput := &autoscaling.CreateAutoScalingGroupInput{
				AutoScalingGroupName: &inputData.ASGName,
				DesiredCapacity:      &inputData.DesiredCount,

				MinSize:           s.MinSize,
				MaxSize:           s.MaxSize,
				CapacityRebalance: s.CapacityRebalance,
				DefaultCooldown:   s.DefaultCooldown,
				PlacementGroup:    s.PlacementGroup,
			}

			if inputData.MinSize != nil {
				createAsgInput.MinSize = inputData.MinSize
			}
			if inputData.MaxSize != nil {
				createAsgInput.MaxSize = inputData.MaxSize
			}

			if len(s.VpcZoneIdentifier) > 0 {
				vpcZoneIdentifiers := strings.Join(s.VpcZoneIdentifier, ",")
				createAsgInput.VPCZoneIdentifier = &vpcZoneIdentifiers
			}

			asgTags := append(s.Tags, inputData.Tags...)
			if len(asgTags) > 0 {
				for _, asgTag := range asgTags {
					createAsgInput.Tags = append(createAsgInput.Tags, &autoscaling.Tag{
						Key:               asgTag.Key,
						Value:             asgTag.Value,
						PropagateAtLaunch: asgTag.PropagateAtLaunch,
					})
				}
			}

			if preferredLaunchTemplateVersion != nil && s.MixedInstancesPolicy == nil {
				// `LaunchTemplate` could be set only when `MixedInstancesPolicy` is not specified.
				stringVersion := fmt.Sprintf("%d", *preferredLaunchTemplateVersion.VersionNumber)
				createAsgInput.LaunchTemplate = &autoscaling.LaunchTemplateSpecification{
					LaunchTemplateId: preferredLaunchTemplateVersion.LaunchTemplateId,
					Version:          &stringVersion,
				}
			}

			if s.MixedInstancesPolicy != nil {
				createAsgInput.MixedInstancesPolicy = &autoscaling.MixedInstancesPolicy{}

				if s.MixedInstancesPolicy.InstanceDistribution != nil {
					createAsgInput.MixedInstancesPolicy.InstancesDistribution = &autoscaling.InstancesDistribution{
						OnDemandAllocationStrategy:          s.MixedInstancesPolicy.InstanceDistribution.OnDemandAllocationStrategy,
						OnDemandBaseCapacity:                s.MixedInstancesPolicy.InstanceDistribution.OnDemandBaseCapacity,
						OnDemandPercentageAboveBaseCapacity: s.MixedInstancesPolicy.InstanceDistribution.OnDemandPercentAboveBaseCapacity,
						SpotAllocationStrategy:              s.MixedInstancesPolicy.InstanceDistribution.SpotAllocationStrategy,
						SpotInstancePools:                   s.MixedInstancesPolicy.InstanceDistribution.SpotInstancePools,
						SpotMaxPrice:                        s.MixedInstancesPolicy.InstanceDistribution.SpotMaxPrice,
					}
				}

				if s.MixedInstancesPolicy.LaunchTemplate != nil {
					createAsgInput.MixedInstancesPolicy.LaunchTemplate = &autoscaling.LaunchTemplate{
						LaunchTemplateSpecification: &autoscaling.LaunchTemplateSpecification{
							LaunchTemplateId:   s.MixedInstancesPolicy.LaunchTemplate.LaunchTemplateSpecification.LaunchTemplateId,
							LaunchTemplateName: s.MixedInstancesPolicy.LaunchTemplate.LaunchTemplateSpecification.LaunchTemplateName,
							Version:            s.MixedInstancesPolicy.LaunchTemplate.LaunchTemplateSpecification.Version,
						},
					}
				}

				if preferredLaunchTemplateVersion != nil {
					// `LaunchTemplate` could be set only when `MixedInstancesPolicy` is not specified.
					stringVersion := fmt.Sprintf("%d", *preferredLaunchTemplateVersion.VersionNumber)
					createAsgInput.MixedInstancesPolicy.LaunchTemplate = &autoscaling.LaunchTemplate{
						LaunchTemplateSpecification: &autoscaling.LaunchTemplateSpecification{
							LaunchTemplateId: preferredLaunchTemplateVersion.LaunchTemplateId,
							Version:          &stringVersion,
						},
					}
				}
			}

			for _, o := range inputData.Overrides {
				if createAsgInput.MixedInstancesPolicy == nil {
					createAsgInput.MixedInstancesPolicy = &autoscaling.MixedInstancesPolicy{}
				}

				if createAsgInput.MixedInstancesPolicy.LaunchTemplate == nil {
					createAsgInput.MixedInstancesPolicy.LaunchTemplate = &autoscaling.LaunchTemplate{}
				}

				override := &autoscaling.LaunchTemplateOverrides{}

				if o.InstanceType != nil {
					override.InstanceType = o.InstanceType
				}

				if o.LaunchTemplateSpecification != nil {
					override.LaunchTemplateSpecification = &autoscaling.LaunchTemplateSpecification{
						LaunchTemplateId:   o.LaunchTemplateSpecification.LaunchTemplateId,
						LaunchTemplateName: o.LaunchTemplateSpecification.LaunchTemplateName,
						Version:            o.LaunchTemplateSpecification.Version,
					}
				}
				createAsgInput.MixedInstancesPolicy.LaunchTemplate.Overrides = append(createAsgInput.MixedInstancesPolicy.LaunchTemplate.Overrides, override)
			}

			s.log.Verbosef("ASG named %s does not exist. So creating it.\n", *createAsgInput.AutoScalingGroupName)
			s.log.Debug("create asg input", createAsgInput)
			createdASG, err := svc.CreateAutoScalingGroup(createAsgInput)
			if err != nil {
				s.log.Error("error while creating ASG", err)
				return
			}
			s.log.Verbosef("created a new ASG: %s", createdASG)
		}

		asgOutput, err = svc.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: []*string{&inputData.ASGName},
			MaxRecords:            &maxRecords,
		})
		if err != nil {
			s.log.Error("error while trying to find autoscaling group", err)
			return
		}

		asg := asgOutput.AutoScalingGroups[0]

		var updateAsg *autoscaling.UpdateAutoScalingGroupInput
		if *asg.DesiredCapacity < inputData.DesiredCount {
			// scale-out
			if s.DisableScaleOut != nil && *s.DisableScaleOut {
				return
			}

			updateAsg = &autoscaling.UpdateAutoScalingGroupInput{
				AutoScalingGroupName: &inputData.ASGName,
				DesiredCapacity:      &inputData.DesiredCount,
			}
		} else if *asg.DesiredCapacity > inputData.DesiredCount {
			// scale-in
			if s.DisableScaleIn != nil && *s.DisableScaleIn {
				return
			}

			updateAsg = &autoscaling.UpdateAutoScalingGroupInput{
				AutoScalingGroupName: &inputData.ASGName,
				DesiredCapacity:      &inputData.DesiredCount,
			}
		}

		if updateAsg != nil {
			s.log.Verbosef("updating asg (%s) to match the desired count: from %d to %d\n", *asg.AutoScalingGroupName, *asg.DesiredCapacity, inputData.DesiredCount)
			_, err = svc.UpdateAutoScalingGroup(updateAsg)
			if err != nil {
				s.log.Error("error trying to update autoscaling group", err)
				return
			}
			s.log.Verbosef("updated asg (%s) to match the desired count: from %d to %d\n", *asg.AutoScalingGroupName, *asg.DesiredCapacity, inputData.DesiredCount)
		}

		s.log.Verbose("end")
	})
	return nil
}

func ParseConfig(k *koanf.Koanf) (scaler.Interface, error) {
	s := Scaler{}
	err := k.Unmarshal("", &s)
	if err != nil {
		return nil, err
	}
	s.namespacedID = fmt.Sprintf("scaler.%s", s.ID)
	fmt.Printf("scaler = %+v \n", s)
	s.log = log.New(s.namespacedID)
	return &s, nil
}
