#!/bin/bash
set -x

# get all the node's ec2 instance ids for the specified chain id
# TODO: fix this and the sharded cluster module to add tags
#chain_node_instance_ids=$(aws ec2 describe-instances --filters "Name=tag:KavaChainId,Values=$CHAIN_ID" | jq -r '[.Reservations | .[] | .Instances | .[] | .InstanceId] | join(" ")')
chain_node_instance_ids='i-0b9b5ebe0d275d4cb'

for chain_node_instance_id in ${chain_node_instance_ids}
do
    autoscaling_group_state=$(aws autoscaling describe-auto-scaling-instances --instance-ids "$chain_node_instance_id" | jq -r '[.AutoScalingInstances | .[].LifecycleState] | join(" ")')
    # Possible states: https://docs.aws.amazon.com/autoscaling/ec2/userguide/ec2-auto-scaling-lifecycle.html
    case "$autoscaling_group_state" in
    InService)
        # place the nodes on standby so they won't get terminated
        # by the autoscaling group during the time
        # they are offline for a deploy / upgrade
        autoscaling_group_name=$(aws autoscaling describe-auto-scaling-instances --instance-ids "$chain_node_instance_id" | jq -r '[.AutoScalingInstances | .[].AutoScalingGroupName] | join(" ")')

        aws autoscaling enter-standby \
            --instance-ids "$chain_node_instance_id" \
            --auto-scaling-group-name "$autoscaling_group_name" \
            --should-decrement-desired-capacity
        ;;
    *)
        echo "instance ($chain_node_instance_id) not in an elgible state ($autoscaling_group_state) for going on standby, skipping"
        ;;
    esac
done
