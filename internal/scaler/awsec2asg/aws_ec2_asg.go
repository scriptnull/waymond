package awsec2asg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/knadh/koanf/v2"
	"github.com/scriptnull/waymond/internal/event"
	"github.com/scriptnull/waymond/internal/log"
	"github.com/scriptnull/waymond/internal/scaler"
)

const Type scaler.Type = "aws_ec2_asg"

type Scaler struct {
	id           string
	namespacedID string
	log          log.Logger
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

	event.B.Subscribe(s.namespacedID, func(data []byte) {
		s.log.Verbose("start")

		s.log.Debugf("data: %+v\n", string(data))

		var inputData struct {
			ASGName      string `json:"asg_name"`
			DesiredCount int64  `json:"desired_count"`
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

		if len(asgOutput.AutoScalingGroups) != 1 {
			s.log.Error("unable to find the autoscaling group", inputData.ASGName)
			return
		}

		_, err = svc.UpdateAutoScalingGroup(&autoscaling.UpdateAutoScalingGroupInput{
			AutoScalingGroupName: &inputData.ASGName,
			DesiredCapacity:      &inputData.DesiredCount,
		})
		if err != nil {
			s.log.Error("error trying to update autoscaling group", err)
			return
		}

		s.log.Verbose("end")
	})
	return nil
}

func ParseConfig(k *koanf.Koanf) (scaler.Interface, error) {
	id := k.String("id")
	if id == "" {
		return nil, errors.New("expected non-empty value for 'id' in aws_ec2_asg scaler")
	}

	s := &Scaler{
		id:           id,
		namespacedID: fmt.Sprintf("scaler.%s", id),
	}
	s.log = log.New(s.namespacedID)
	return s, nil
}
