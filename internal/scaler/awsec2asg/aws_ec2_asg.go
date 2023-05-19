package awsec2asg

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
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

	AllowCreate          bool                              `koanf:"allow_create"`
	MinSize              *int64                            `koanf:"min_size"`
	MaxSize              *int64                            `koanf:"max_size"`
	CapacityRebalance    *bool                             `koanf:"capacity_rebalance"`
	DefaultCooldown      *int64                            `koanf:"default_cooldown"`
	VpcZoneIdentifier    []string                          `koanf:"vpc_zone_identifier"`
	PlacementGroup       *string                           `koanf:"placement_group"`
	Tags                 []ASGTag                          `koanf:"tags"`
	MixedInstancesPolicy *autoscaling.MixedInstancesPolicy `koanf:"mixed_instances_policy"`
}

type ASGTag struct {
	Key               *string `koanf:"key"`
	Value             *string `koanf:"value"`
	PropagateAtLaunch *bool   `koanf:"propagate_at_launch"`
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

	event.B.Subscribe(fmt.Sprintf("%s.input", s.namespacedID), func(data []byte) {
		s.log.Verbose("start")

		s.log.Debugf("data: %+v\n", string(data))

		var inputData struct {
			ASGName      string   `json:"asg_name"`
			DesiredCount int64    `json:"desired_count"`
			Tags         []ASGTag `json:"tags"`
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

				MinSize:              s.MinSize,
				MaxSize:              s.MaxSize,
				CapacityRebalance:    s.CapacityRebalance,
				DefaultCooldown:      s.DefaultCooldown,
				PlacementGroup:       s.PlacementGroup,
				MixedInstancesPolicy: s.MixedInstancesPolicy,
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

			s.log.Debug("create asg input", createAsgInput)
			// createdASG, err := svc.CreateAutoScalingGroup(createAsgInput)
			// if err != nil {
			// 	s.log.Error("error while creating ASG", err)
			// 	return
			// }
			// s.log.Debug("created asg", createdASG)
			// s.log.Verbosef("successfully created a new ASG: %s", createdASG.String())
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
