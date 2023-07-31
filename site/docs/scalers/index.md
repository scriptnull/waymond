# Scalers

Scalers are the components which perform autoscaling operation. They access the target APIs (like docker or AWS API) to achieve a desired state in those systems.

| type           | status                   | description                                                                                                         |
| -------------- | ------------------------ | ------------------------------------------------------------------------------------------------------------------- |
| docker         | Available                | Autoscales docker containers                                                                                        |
| docker_compose | Looking for contribution | Autoscales containers in a docker compose setup                                                                     |
| noop           | Available                | `noop` stands for "No-Operation". This is mainly for debugging what data is being received by a problematic scaler. |
| aws_ec2        | Looking for contribution | Autoscales AWS EC2 machines                                                                                         |
| aws_ec2_fleet  | Looking for contribution | Autoscales AWS EC2 machines via [EC2 Fleet](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-fleet.html)     |
| aws_ec2_asg    | Available                | Autoscales AWS EC2 machines via Autoscaling groups                                                                  |

Propose a new scaler [here](https://github.com/scriptnull/waymond/issues/new).
