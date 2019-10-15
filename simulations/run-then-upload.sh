#!/bin/bash

# This requires AWS access keys envs to be set (ie AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)
# These need to be generated from the AWS console.

# For commands passed to the docker container, the working directory is /root/kava (which is the blockchain git repo).


# Parse Input Args
# get simulation type
simType=$1
# get seed
startingSeed=$2
# compute the seed from the starting and the job index
# add two nums together, hence the $(()), and use 0 as the default value for array index, hence the ${:-} syntax
seed=$(($startingSeed+${AWS_BATCH_JOB_ARRAY_INDEX:-0}))
# get sim parameters
numBlocks=$3
blockSize=$4


# Run The Sim
# record cli arguments in the log file (in json in case we need to parse this) and stdout (https://stackoverflow.com/questions/418896/how-to-redirect-output-to-a-file-and-stdout)
printf "{\"simType\": \"%s\", \"startingSeed\": %s, \"seed\": %s, \"numBlocks\": %s, \"blockSize\": %s}\n" $simType $startingSeed $seed $numBlocks $blockSize | tee out.log
# run sim and redirect stdout and stderr to a file)
go test ./app -run $simType -Enabled=true -NumBlocks=$numBlocks -BlockSize=$blockSize -Commit=true -Period=5 -Seed=$seed -v -timeout 24h >> out.log 2>&1
# get the exit code to determine how to upload results
simExitStatus=$?
if [ $simExitStatus -eq 0 ];then
   echo "simulation passed"
   simResult="pass"
else
   echo "simulation failed"
   simResult="fail"
fi


# Upload Sim Results To S3
# read in the job id, using a default value if not set
jobID=${AWS_BATCH_JOB_ID:-"testJobID:"}
# job id format is "job-id:array-job-index", this removes trailing colon (and array index if present) https://stackoverflow.com/questions/3045493/parse-string-with-bash-and-extract-number
jobID=$(echo $jobID | sed 's/\(.*\):\d*/\1/')

# create the filename from the array job index (which won't be set if this is a normal job)
fileName=$(printf "out%05d.log" $AWS_BATCH_JOB_ARRAY_INDEX)

aws s3 cp out.log s3://simulations-1/$SIM_NAME$jobID/$simResult/$fileName