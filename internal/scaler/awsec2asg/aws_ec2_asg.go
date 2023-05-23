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
	LaunchTemplateId   *string                        `koanf:"launch_template_id" json:"launch_template_id"`
	LaunchTemplateName *string                        `koanf:"launch_template_name" json:"launch_template_name"`
	Version            *string                        `koanf:"version" json:"version"`
	CreateIfNotExists  *bool                          `koanf:"create_if_not_exists" json:"create_if_not_exists"`
	Spec               *ec2.CreateLaunchTemplateInput `koanf:"spec" json:"spec"`
}

type LaunchTemplateOverrides struct {
	// TODO: too much of struct nesting in InstanceRequirements
	// so add it here when we actually need it.
	// InstanceRequirements        *InstanceRequirements        `koanf:"instance_requirements"`

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
			ASGName      string                    `json:"asg_name"`
			DesiredCount int64                     `json:"desired_count"`
			Tags         []ASGTag                  `json:"tags"`
			Overrides    []LaunchTemplateOverrides `json:"launch_template_overrides"`
		}

		err := json.Unmarshal(data, &inputData)
		if err != nil {
			s.log.Errorf("error: %s \n", err)
			return
		}

		var maxRecords int64 = 1
		asg, err := svc.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: []*string{&inputData.ASGName},
			MaxRecords:            &maxRecords,
		})
		if err != nil {
			s.log.Error("error while trying to find autoscaling group", err)
			return
		}

		s.log.Debug("asg", asg)

		if len(asg.AutoScalingGroups) == 0 {
			// autoscaling group is absent in EC2
			// so create one if allow_create is set to true
			// else error out and return

			if !s.AllowCreate {
				s.log.Error("ASG not found. Please set `allow_create` to true if you would like to create it via waymond")
				return
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

			if s.MixedInstancesPolicy != nil {
				createAsgInput.MixedInstancesPolicy = &autoscaling.MixedInstancesPolicy{
					InstancesDistribution: &autoscaling.InstancesDistribution{
						OnDemandAllocationStrategy:          s.MixedInstancesPolicy.InstanceDistribution.OnDemandAllocationStrategy,
						OnDemandBaseCapacity:                s.MixedInstancesPolicy.InstanceDistribution.OnDemandBaseCapacity,
						OnDemandPercentageAboveBaseCapacity: s.MixedInstancesPolicy.InstanceDistribution.OnDemandPercentAboveBaseCapacity,
						SpotAllocationStrategy:              s.MixedInstancesPolicy.InstanceDistribution.SpotAllocationStrategy,
						SpotInstancePools:                   s.MixedInstancesPolicy.InstanceDistribution.SpotInstancePools,
						SpotMaxPrice:                        s.MixedInstancesPolicy.InstanceDistribution.SpotMaxPrice,
					},
					LaunchTemplate: &autoscaling.LaunchTemplate{
						LaunchTemplateSpecification: &autoscaling.LaunchTemplateSpecification{
							LaunchTemplateId:   s.MixedInstancesPolicy.LaunchTemplate.LaunchTemplateSpecification.LaunchTemplateId,
							LaunchTemplateName: s.MixedInstancesPolicy.LaunchTemplate.LaunchTemplateSpecification.LaunchTemplateName,
							Version:            s.MixedInstancesPolicy.LaunchTemplate.LaunchTemplateSpecification.Version,
						},
					},
				}
			}

			for _, o := range inputData.Overrides {
				s.log.Debugf("o = %+v \n", o)

				if o.LaunchTemplateSpecification.CreateIfNotExists != nil && *o.LaunchTemplateSpecification.CreateIfNotExists {
					if o.LaunchTemplateSpecification.Spec == nil {
						s.log.Error("create_if_not_exists option for launch template expects spec object to be present")
						return
					}

					// check if the launch template exists
					lts, err := ec2svc.DescribeLaunchTemplates(&ec2.DescribeLaunchTemplatesInput{
						LaunchTemplateNames: []*string{o.LaunchTemplateSpecification.LaunchTemplateName},
					})
					if err != nil {
						if !strings.Contains(err.Error(), "InvalidLaunchTemplateName.NotFoundException") {
							s.log.Error("error while checking for launch template", err)
							continue
						}
					}

					if lts != nil && len(lts.LaunchTemplates) > 0 {
						s.log.Debugf("launch template named %s exists, so avoiding the creation of it.\n", *o.LaunchTemplateSpecification.LaunchTemplateName)
						continue
					}

					// create a launch template
					s.log.Verbosef("launch template named %s does not exist, so creating it.\n", *o.LaunchTemplateSpecification.LaunchTemplateName)
					o.LaunchTemplateSpecification.Spec.LaunchTemplateName = o.LaunchTemplateSpecification.LaunchTemplateName
					s.log.Debug("launch template spec", o.LaunchTemplateSpecification.Spec)
					ltOut, err := ec2svc.CreateLaunchTemplate(o.LaunchTemplateSpecification.Spec)
					if err != nil {
						s.log.Error("error while creating launch template", err)
						continue
					}

					s.log.Verbosef("created a new launch template: %s\n", ltOut)
				}
			}

			for _, o := range inputData.Overrides {
				createAsgInput.MixedInstancesPolicy.LaunchTemplate.Overrides = append(createAsgInput.MixedInstancesPolicy.LaunchTemplate.Overrides, &autoscaling.LaunchTemplateOverrides{
					InstanceType: o.InstanceType,
					LaunchTemplateSpecification: &autoscaling.LaunchTemplateSpecification{
						LaunchTemplateId:   o.LaunchTemplateSpecification.LaunchTemplateId,
						LaunchTemplateName: o.LaunchTemplateSpecification.LaunchTemplateName,
						Version:            o.LaunchTemplateSpecification.Version,
					},
				})
			}

			s.log.Verbosef("ASG named %s does not exist. So creating it.\n", *createAsgInput.AutoScalingGroupName)
			s.log.Debug("create asg input", createAsgInput)
			createdASG, err := svc.CreateAutoScalingGroup(createAsgInput)
			if err != nil {
				s.log.Error("error while creating ASG", err)
				return
			}
			s.log.Verbosef("created a new ASG: %s", createdASG)
			return
		}

		// if len(asgOutput.AutoScalingGroups) != 1 {
		// 	s.log.Error("unable to find the autoscaling group", inputData.ASGName)
		// 	return
		// }

		// _, err = svc.UpdateAutoScalingGroup(&autoscaling.UpdateAutoScalingGroupInput{
		// 	AutoScalingGroupName: &inputData.ASGName,
		// 	DesiredCapacity:      &inputData.DesiredCount,
		// })
		// if err != nil {
		// 	s.log.Error("error trying to update autoscaling group", err)
		// 	return
		// }

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
